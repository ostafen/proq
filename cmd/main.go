package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"

	"github.com/ostafen/proq/pkg/metric"
	"github.com/ostafen/proq/pkg/store"
	wg "github.com/ostafen/proq/pkg/widgets"
)

type MetricsDash struct {
	plot   *wg.MetricPlot
	list   *wg.MetricList
	hist   *wg.Histogram
	prompt *wg.Prompt

	plotHeightPerc float64
	plotWidthPerc  float64
}

func (dash *MetricsDash) Resize(width, height int) {
	dash.plot.HorizontalScale = (width - 4) / (dash.plot.NumSamples - 1)

	const barHeight = 3

	dash.plot.SetRect(0, 0, int(float64(width)*dash.plotWidthPerc), int(float64(height)*dash.plotHeightPerc))
	dash.list.SetRect(0, int(float64(height)*dash.plotHeightPerc), width, height-barHeight)

	dash.prompt.SetRect(0, height-barHeight, width, height)
}

func (s *App) Run() {
	dash := s.dash

	resize := func() {
		width, height := ui.TerminalDimensions()

		dash.Resize(width, height)
		ui.Render(dash.list, dash.plot, dash.prompt)
	}

	resize()

	start := time.Now()
	ticker := time.NewTicker(s.fetchInterval)

	uiEvents := ui.PollEvents()
	for {
		select {
		case <-ticker.C:
			s.fetch()
		case e := <-uiEvents:
			if e.Type == ui.KeyboardEvent {
				s.dash.OnKeyPressed(e.ID)
			} else if e.Type == ui.ResizeEvent {
				resize()
			}
		case v := <-s.ch:
			n := dash.plot.NumSamples / 2

			if len(s.dash.plot.Data[0])+1 > dash.plot.NumSamples {
				s.dash.plot.Data[0] = append(s.dash.plot.Data[0][n+1:], v)
				s.dash.plot.Refresh(time.Since(start))
			} else {
				s.dash.plot.Data[0] = append(s.dash.plot.Data[0], v)
			}
			ui.Render(s.dash.plot, s.dash.prompt)
		}
	}
}

type App struct {
	displayWindow time.Duration

	stream        *store.Stream
	ch            chan float64
	url           string
	fetchInterval time.Duration

	dash  *MetricsDash
	store *store.MetricStore
}

func (app *App) renderMetric(m wg.MetricInfo) {
	if app.stream != nil {
		app.stream.Close()
		close(app.ch)
		app.ch = make(chan float64, 1)
	}

	mk := metric.MetricKey{
		Name:   m.Name,
		Labels: m.Labels,
	}

	if m.IsHist {
		app.renderHistogram(mk)
	} else {
		app.renderGenericMetric(mk)
	}
}

func (app *App) renderHistogram(m metric.MetricKey) {
	h := app.store.GetHist(m)

	width, height := ui.TerminalDimensions()

	app.dash.hist = wg.NewHistogram(h, width)

	app.dash.hist.SetRect(0, 0, int(float64(width)), int(float64(height)*0.7))
	ui.Render(app.dash.hist.BarChart)
}

func (app *App) renderGenericMetric(m metric.MetricKey) {
	st := app.store
	dash := app.dash

	var samples []float64
	n := st.Samples(m, func(f float64) {
		samples = append(samples, f)
	})

	if len(samples) > 0 {
		if len(dash.plot.Data) == 0 {
			dash.plot.Data = [][]float64{samples}
		} else {
			dash.plot.Data[0] = samples
		}
	}

	app.stream = st.Bind(m, app.ch, n)
	dash.plot.Title = m.Name
	ui.Render(app.dash.plot)
}

func (dash *MetricsDash) OnKeyPressed(key string) bool {
	drawables := make([]ui.Drawable, 0)

	if dash.prompt.OnKeyPressed(key) {
		drawables = append(drawables, dash.prompt)
	}

	if dash.list.OnKeyPressed(key) {
		drawables = append(drawables, dash.list)
	}

	ui.Render(drawables...)
	return false
}

func (app *App) cmdsHandlers() map[string]wg.CmdHandler {
	return map[string]wg.CmdHandler{
		"q": app.quit,
		"f": app.filter,
	}
}

func (app *App) filter(_ string, args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("no filter specified")
	}
	return app.dash.list.Filter(args[0])
}

func (s *App) quit(_ string, args ...string) error {
	ui.Close()

	os.Exit(0)
	return nil
}

func (s *App) Start() {
	s.Run()
}

func (s *App) fetch() {
	histo, rawMetrics, err := s.fetchMetrics(s.url)
	if err != nil {
		return
	}

	for _, m := range rawMetrics {
		s.store.Update(&m)
	}

	s.store.UpdateHistogram(histo)

	metrics := make([]wg.MetricInfo, len(rawMetrics)+len(histo))
	for i, m := range rawMetrics {
		metrics[i] = wg.MetricInfo{
			Name:   m.Name,
			Labels: m.Labels,
			IsHist: false,
		}
	}

	i := 0
	for _, m := range histo {
		metrics[len(rawMetrics)+i] = wg.MetricInfo{
			Name:   m.Name,
			Labels: m.Labels,
			IsHist: true,
		}
		i++
	}

	sort.Slice(metrics, func(i, j int) bool {
		m1 := metrics[i]
		m2 := metrics[j]

		if m1.IsHist != m2.IsHist {
			return m1.IsHist && !m2.IsHist
		}

		res := strings.Compare(m1.Name, m2.Name)
		return res < 0
	})

	s.dash.list.Set(metrics)
}

func (s *App) fetchMetrics(url string) (map[string]metric.Histogram, []metric.RawMetric, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("error fetching metrics: %v", err)
	}
	defer resp.Body.Close()

	rawMetrics := make([]metric.RawMetric, 0, 100)

	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		m, err := metric.ParseMetricLine(line)
		if err != nil {
			fmt.Printf("unable to parse line: %s\n", line)
			continue
		}

		rawMetrics = append(rawMetrics, m)
	}

	histograms, rem := metric.ParseHistogram(rawMetrics)
	return histograms, rem, sc.Err()
}

const (
	DefaultDisplayWindow = time.Minute
	DefaultPollInterval  = 1 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("no url specified")
		os.Exit(1)
	}

	url := os.Args[1]
	os.Args = os.Args[1:]

	displayWindow := flag.Duration("window", DefaultDisplayWindow, "time size of displayed window")
	pollInterval := flag.Duration("refresh-interval", DefaultPollInterval, "the frequency the metric endpoint is queries")

	flag.Parse()

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	maxSamples := int(*displayWindow/(*pollInterval)) + 1
	metricStore := store.NewMetricStore(maxSamples)

	dash := &MetricsDash{
		prompt: wg.NewPrompt(),
		plot: wg.NewMetricPlot(
			5,
			*pollInterval,
			*displayWindow,
		),
		plotHeightPerc: 0.7,
		plotWidthPerc:  1,
	}

	app := &App{
		displayWindow: *displayWindow,
		fetchInterval: *pollInterval,
		ch:            make(chan float64, 1),
		url:           url,
		store:         metricStore,
		dash:          dash,
	}

	dash.list = wg.NewMetricList(app.renderMetric)
	dash.prompt.SetHandlers(app.cmdsHandlers())

	app.Start()
}
