package widgets

import (
	"time"

	ui "github.com/ostafen/termui/v3"
	"github.com/ostafen/termui/v3/widgets"
)

type MetricPlot struct {
	*widgets.Plot

	NumSamples   int
	numTicks     int
	tickInterval time.Duration
	currMaxTick  time.Duration

	start time.Time
}

func NewMetricPlot(
	numTicks int,
	sampleRate time.Duration,
	windowInterval time.Duration,
) *MetricPlot {
	plot := widgets.NewPlot()
	plot.Title = "Metric Data"
	plot.Data = [][]float64{}
	plot.AxesColor = ui.ColorBlack
	plot.LineColors[0] = ui.ColorGreen
	plot.Marker = widgets.MarkerBraille

	maxWindowSamples := int(windowInterval/sampleRate) + 1
	if numTicks > maxWindowSamples {
		numTicks = maxWindowSamples
	}

	tickInterval := windowInterval / time.Duration(numTicks)

	labels := make([]string, numTicks+1)
	for i := range labels {
		axisLabel := tickInterval * time.Duration(i)
		labels[i] = axisLabel.String()
	}
	plot.DataLabels = labels

	return &MetricPlot{
		Plot:         plot,
		NumSamples:   maxWindowSamples,
		numTicks:     numTicks,
		tickInterval: tickInterval,
		currMaxTick:  windowInterval,
		start:        time.Now(),
	}
}

func (p *MetricPlot) Update(sample float64) {
	n := p.NumSamples / 2

	if len(p.Data[0])+1 > p.NumSamples {
		p.Data[0] = append(p.Data[0][n+1:], sample)
		p.Refresh(time.Since(p.start))
	} else {
		p.Data[0] = append(p.Data[0], sample)
	}

	p.MaxVal = max(p.Data[0])

	ui.Render(p)
}

func max(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func (p *MetricPlot) Refresh(elapsedTime time.Duration) {
	p.currMaxTick += (p.tickInterval * time.Duration(p.numTicks)) / 2

	labels := make([]string, p.numTicks+1)
	for i := range labels {
		axisLabel := p.currMaxTick - p.tickInterval*time.Duration(i)
		labels[len(labels)-i-1] = axisLabel.String()
	}
	p.Plot.DataLabels = labels
	ui.Render(p)
}
