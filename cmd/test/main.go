package main

import (
	"log"
	"math"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// HistogramData represents parsed Prometheus histogram buckets.
type HistogramData struct {
	BucketLabels []string  // Bucket upper bounds as labels (e.g., "0.1", "0.5", "1.0")
	BucketValues []float64 // Counts of observations in each bucket
}

// Sample histogram data (simulating Prometheus metrics parsing)
func getSampleHistogramData() HistogramData {
	return HistogramData{
		BucketLabels: []string{"0.1", "0.5", "1", "2", "+Inf"},
		BucketValues: []float64{1, 1, 1, 10, 0},
	}
}

func renderHistogram(histogram HistogramData) {
	if err := ui.Init(); err != nil {
		log.Fatalf("Failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Create a bar chart
	barChart := widgets.NewBarChart()
	barChart.Title = "Histogram: Request Duration (Seconds)"
	barChart.Data = histogram.BucketValues
	barChart.Labels = histogram.BucketLabels
	barChart.BarWidth = 100
	barChart.BarGap = 2
	barChart.MaxVal = getMax(histogram.BucketValues) // Set max Y-axis value
	barChart.BorderStyle.Fg = ui.ColorCyan

	// Set grid layout
	//grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	barChart.SetRect(0, 0, termWidth, termHeight)
	//grid.Set(ui.NewRow(1.0, barChart))

	// Render loop
	ui.Render(barChart)
	uiEvents := ui.PollEvents()
	for e := range uiEvents {
		if e.Type == ui.KeyboardEvent && strings.ToLower(e.ID) == "q" {
			break // Quit on 'q'
		}
	}
}

// Helper function to get the max value for Y-axis scaling
func getMax(values []float64) float64 {
	max := float64(0)
	for _, v := range values {
		max = math.Max(max, v)
	}
	return max
}

func main() {
	histogram := getSampleHistogramData()
	renderHistogram(histogram)
}
