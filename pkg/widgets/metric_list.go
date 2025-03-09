package widgets

import (
	"fmt"
	"regexp"
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

func (mi *MetricInfo) Key() metric.MetricKey {
	return metric.MetricKey{
		Name:   mi.Name,
		Labels: mi.Labels,
	}
}

type MetricList struct {
	selectedRow int

	allMetrics map[string]struct{}

	metrics          []MetricInfo
	displayedMetrics []MetricInfo

	onMetricSelected func(m MetricInfo)

	*widgets.List
}

func NewMetricList(onMetricSelected func(m MetricInfo)) *MetricList {
	list := widgets.NewList()
	list.Title = "Metrics"
	list.Rows = nil
	list.TextStyle = ui.NewStyle(ui.ColorYellow)

	return &MetricList{
		allMetrics:       make(map[string]struct{}),
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
	if l.selectedRow < 0 || l.selectedRow >= len(l.displayedMetrics) {
		return false
	}

	m := l.displayedMetrics[l.selectedRow]
	l.onMetricSelected(m)
	return true
}

func (l *MetricList) ShowHistograms() {
	metrics := make([]MetricInfo, 0, len(l.allMetrics))
	for _, m := range l.metrics {
		if m.IsHist {
			metrics = append(metrics, m)
		}
	}

	l.displayedMetrics = metrics
	l.RenderList()
}

func wildcardToRegex(pattern string) string {
	replacer := strings.NewReplacer(
		".", "\\.",
		"+", "\\+",
		"(", "\\(",
		")", "\\)",
		"{", "\\{",
		"}", "\\}",
		"|", "\\|",
		"^", "\\^",
		"$", "\\$",
	)

	pattern = replacer.Replace(pattern)
	pattern = strings.ReplaceAll(pattern, "*", ".*") // * -> .*
	pattern = strings.ReplaceAll(pattern, "?", ".")  // ? -> .

	return "^" + pattern + "$"
}

func (l *MetricList) Filter(pattern string) error {
	exp, err := regexp.Compile(wildcardToRegex(pattern))
	if err != nil {
		return err
	}

	if pattern == "" {
		l.displayedMetrics = l.metrics
		return fmt.Errorf("empty filter")
	}

	metrics := make([]MetricInfo, 0, len(l.metrics))
	for _, m := range l.metrics {
		if exp.MatchString(m.Name) {
			metrics = append(metrics, m)
		}
	}

	if len(metrics) == 0 {
		return fmt.Errorf("\"%s\": no metric matches the specified filter", pattern)
	}
	l.displayedMetrics = metrics

	l.RenderList()
	return nil
}

func (l *MetricList) Reset() {
	l.displayedMetrics = l.metrics
	l.RenderList()
}

func (l *MetricList) AddMetrics(metrics []MetricInfo) {
	for _, m := range metrics {
		mk := m.Key()
		s := mk.String()
		_, has := l.allMetrics[s]
		if !has {
			l.metrics = append(l.metrics, m)
		}
		l.allMetrics[s] = struct{}{}
	}

	if l.displayedMetrics == nil {
		l.displayedMetrics = l.metrics
	}

	l.RenderList()
}

func (l *MetricList) RenderList() {
	rows := make([]string, len(l.displayedMetrics))
	for i, m := range l.displayedMetrics {
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

	l.Title = fmt.Sprintf("Metrics (%d/%d)", len(rows), len(l.allMetrics))

	ui.Render(l.List)
}
