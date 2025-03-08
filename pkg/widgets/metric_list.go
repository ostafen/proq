package widgets

import (
	"fmt"
	"strings"

	ui "github.com/ostafen/termui/v3"

	"github.com/ostafen/proq/pkg/metric"
	"github.com/ostafen/termui/v3/widgets"
)

type MetricInfo struct {
	Name   string
	Labels []metric.Label
	IsHist bool
}

type MetricList struct {
	selectedRow int

	allMetrics      []MetricInfo
	filteredMetrics []MetricInfo

	onMetricSelected func(m MetricInfo)

	*widgets.List
}

func NewMetricList(onMetricSelected func(m MetricInfo)) *MetricList {
	list := widgets.NewList()
	list.Title = "Metrics"
	list.Rows = nil
	list.TextStyle = ui.NewStyle(ui.ColorYellow)

	return &MetricList{
		List:             list,
		onMetricSelected: onMetricSelected,
	}
}

func (l *MetricList) Resize(width int, height int) {}

func (l *MetricList) OnKeyPressed(key string) bool {
	switch key {
	case "<Up>":
		return l.scroll(-1)
	case "<Down>":
		return l.scroll(1)
	}
	return false
}

func (l *MetricList) scroll(direction int) bool {
	if l.selectedRow+direction < 0 || l.selectedRow+direction >= len(l.Rows) {
		return false
	}

	l.selectedRow += direction
	l.selectedRow = (l.selectedRow + len(l.Rows)) % len(l.Rows)

	if direction < 0 {
		l.ScrollUp()
	} else {
		l.ScrollDown()
	}
	return l.selectMetric()
}

func (l *MetricList) selectMetric() bool {
	if l.selectedRow < 0 || l.selectedRow >= len(l.filteredMetrics) {
		return false
	}

	m := l.filteredMetrics[l.selectedRow]
	l.onMetricSelected(m)
	return true
}

func (l *MetricList) Filter(filter string) error {
	if filter == "" {
		l.filteredMetrics = l.allMetrics
		return fmt.Errorf("empty filter")
	}

	metrics := make([]MetricInfo, 0, len(l.allMetrics))
	for _, m := range l.allMetrics {
		if strings.Contains(m.Name, filter) {
			metrics = append(metrics, m)
		}
	}
	if len(metrics) == 0 {
		return fmt.Errorf("\"%s\": no metric matches the specified filter", filter)
	}
	l.filteredMetrics = metrics

	l.RenderList()
	return nil
}

func (l *MetricList) Set(list []MetricInfo) {
	l.allMetrics = list
	if l.filteredMetrics == nil {
		l.filteredMetrics = list
	}

	l.RenderList()
}

func (l *MetricList) RenderList() {
	rows := make([]string, len(l.filteredMetrics))
	for i, m := range l.filteredMetrics {
		mk := metric.MetricKey{
			Name:   m.Name,
			Labels: m.Labels,
		}

		rows[i] = "- " + mk.String()
	}

	if len(rows) == 0 {
		l.Rows = []string{}
	} else {
		l.Rows = rows
	}

	l.Title = fmt.Sprintf("Metrics (%d)", len(rows))

	ui.Render(l.List)
}
