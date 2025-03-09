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

	ui "github.com/ostafen/termui/v3"

	"github.com/ostafen/proq/pkg/metric"
	"github.com/ostafen/proq/pkg/store"
	wg "github.com/ostafen/proq/pkg/widgets"
)

type App struct {
	displayWindow time.Duration

	stream        *store.Stream
	ch            chan float64
	url           string
	fetchInterval time.Duration

	dash  *wg.MetricsDash
	store *store.MetricStore
}

func (s *App) Start() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	s.dash.Resize()

	ticker := time.NewTicker(s.fetchInterval)
	uiEvents := ui.PollEvents()
	for {
		select {
		case <-ticker.C:
			s.fetch()
		case e := <-uiEvents:
			s.handleUIEvent(e)
		case v := <-s.ch:
			s.dash.Plot.Update(v)
		}
	}
}

func (app *App) handleUIEvent(e ui.Event) {
	switch e.Type {
	case ui.KeyboardEvent:
		app.dash.OnKeyPressed(e.ID)
	case ui.ResizeEvent:
		app.dash.Resize()
	}
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

	app.dash.Hist = wg.NewHistogram(h, width)

	app.dash.Hist.SetRect(0, 0, int(float64(width)), int(float64(height)*0.7))
	ui.Render(app.dash.Hist.BarChart)
}

func (app *App) renderGenericMetric(m metric.MetricKey) {
	st := app.store
	dash := app.dash

	var samples []float64
	n := st.Samples(m, func(f float64) {
		samples = append(samples, f)
	})

	if len(samples) > 0 {
		if len(dash.Plot.Data) == 0 {
			dash.Plot.Data = [][]float64{samples}
		} else {
			dash.Plot.Data[0] = samples
		}
	}

	app.stream = st.Bind(m, app.ch, n)
	dash.Plot.Title = m.Name
	ui.Render(app.dash.Plot)
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
	return app.dash.List.Filter(args[0])
}

func (s *App) quit(_ string, args ...string) error {
	ui.Close()

	os.Exit(0)
	return nil
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

	s.dash.List.Set(metrics)
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

	maxSamples := int(*displayWindow/(*pollInterval)) + 1
	metricStore := store.NewMetricStore(maxSamples)

	dash := &wg.MetricsDash{
		Prompt: wg.NewPrompt(),
		Plot: wg.NewMetricPlot(
			*pollInterval,
			*displayWindow,
		),
	}

	app := &App{
		displayWindow: *displayWindow,
		fetchInterval: *pollInterval,
		ch:            make(chan float64, 1),
		url:           url,
		store:         metricStore,
		dash:          dash,
	}

	dash.List = wg.NewMetricList(app.renderMetric)
	dash.Prompt.SetHandlers(app.cmdsHandlers())

	app.Start()
}
