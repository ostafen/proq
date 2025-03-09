package widgets

import (
	"time"

	ui "github.com/ostafen/termui/v3"
)

const (
	HeightRatio = 0.7
	WidthRatio  = 1
)

type MetricsDash struct {
	Plot   *MetricPlot
	List   *MetricList
	Hist   *Histogram
	Prompt *Prompt
}

func NewMetricDash(
	pollInterval time.Duration,
	displayInterval time.Duration,
) *MetricsDash {
	return &MetricsDash{
		Prompt: NewPrompt(),
		Plot: NewMetricPlot(
			pollInterval,
			displayInterval,
		),
	}
}

const yAxisLabelsWidth = 4 // from termui

func (dash *MetricsDash) Resize() {
	width, height := ui.TerminalDimensions()

	dash.Plot.HorizontalScale = float64(width-yAxisLabelsWidth-1) / float64(dash.Plot.NumSamples-1)
	dash.Plot.CurrWidth = width

	const barHeight = 3

	dash.Plot.SetRect(0, 0, int(float64(width)*WidthRatio), int(float64(height)*HeightRatio))
	dash.List.SetRect(0, int(float64(height)*HeightRatio), width, height-barHeight)

	dash.Prompt.SetRect(0, height-barHeight, width, height)

	dash.Render()
}

func (dash *MetricsDash) Render() {
	ui.Render(dash.List, dash.Plot, dash.Prompt)
}

func (dash *MetricsDash) OnKeyPressed(key string) bool {
	drawables := make([]ui.Drawable, 0)

	if dash.Prompt.OnKeyPressed(key) {
		drawables = append(drawables, dash.Prompt)
	}

	if dash.List.OnKeyPressed(key) {
		drawables = append(drawables, dash.List)
	}

	ui.Render(drawables...)
	return false
}

func (dash *MetricsDash) SetMetricList(metrics []MetricInfo) {
	dash.List.AddMetrics(metrics)
}

func (dash *MetricsDash) ShowHistograms() {
	dash.List.ShowHistograms()
}

func (dash *MetricsDash) FilterMetrics(filter string) error {
	return dash.List.Filter(filter)
}

func (dash *MetricsDash) ResetMetrics() {
	dash.List.Reset()
}
