package widgets

import (
	"math"
	"strconv"

	ui "github.com/gizak/termui/v3"

	"github.com/gizak/termui/v3/widgets"
	"github.com/ostafen/proq/pkg/metric"
)

type Histogram struct {
	*widgets.BarChart
}

func NewHistogram(hist *metric.Histogram, width int) *Histogram {
	const barWidth = 7

	// if there is not enough space to render all the bins,
	// merge the last bins to the +Inf bin.
	barGap := (width - barWidth*(len(hist.Bins))) / len(hist.Bins)
	for i := len(hist.Bins) - 1; i > 0; i-- {
		if barGap >= 2 {
			break
		}

		hist.Bins[i-1].Count += hist.Bins[i].Count
		hist.Bins[i-1].Value = math.Inf(1)
		hist.Bins = hist.Bins[:len(hist.Bins)-1]

		barGap = (width - barWidth*(len(hist.Bins))) / len(hist.Bins)
	}

	bucketValues := make([]float64, len(hist.Bins))
	bucketLabels := make([]string, len(hist.Bins))

	var maxVal float64 = hist.Bins[0].Value
	for i, b := range hist.Bins {
		bucketLabels[i] = strconv.FormatFloat(b.Value, 'f', 0, 64)
		bucketValues[i] = float64(b.Count)

		if float64(b.Count) > maxVal {
			maxVal = float64(b.Count)
		}
	}

	barChart := widgets.NewBarChart()
	barChart.Title = hist.Name
	barChart.Data = bucketValues
	barChart.Labels = bucketLabels
	barChart.BarWidth = barWidth

	barChart.BarGap = barGap
	barChart.PaddingLeft = barChart.BarGap / 2

	barChart.MaxVal = maxVal
	barChart.BorderStyle.Fg = ui.ColorWhite

	barChart.BarColors = make([]ui.Color, len(bucketValues))
	for i := range barChart.BarColors {
		barChart.BarColors[i] = ui.ColorGreen
	}

	barChart.LabelStyles = make([]ui.Style, len(bucketLabels))
	for i := range barChart.LabelStyles {
		barChart.LabelStyles[i] = ui.NewStyle(ui.ColorWhite)
	}

	barChart.NumStyles = make([]ui.Style, len(bucketValues))
	for i := range barChart.NumStyles {
		barChart.NumStyles[i] = ui.NewStyle(ui.ColorBlack)
	}

	return &Histogram{
		BarChart: barChart,
	}
}
