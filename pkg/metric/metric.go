package metric

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var ErrInvalidMetricLine = errors.New("invalid metric line")

type Label struct {
	Name  string
	Value string
}

func (l *Label) String() string {
	return fmt.Sprintf(`%s="%s"`, l.Name, l.Value)
}

type MetricKey struct {
	Name   string
	Labels []Label
}

func (mk *MetricKey) String() string {
	var sb strings.Builder

	sb.WriteString(mk.Name)

	if len(mk.Labels) > 0 {
		sb.WriteString("{")

		for _, l := range mk.Labels[:len(mk.Labels)-1] {
			sb.WriteString(l.String() + ", ")
		}

		l := mk.Labels[len(mk.Labels)-1]
		sb.WriteString(l.String())

		sb.WriteString("}")
	}

	return sb.String()
}

type RawMetric struct {
	Name   string
	Labels []Label
	Value  float64
}

func (m *RawMetric) Find(name string) string {
	for _, l := range m.Labels {
		if l.Name == name {
			return l.Value
		}
	}
	return ""
}

func (m *RawMetric) Remove(name string) ([]Label, bool) {
	if len(m.Labels) == 0 {
		return nil, false
	}

	labels := make([]Label, 0, len(m.Labels))
	for _, l := range m.Labels {
		if l.Name != name {
			labels = append(labels, l)
		}
	}
	return labels, len(labels) < len(m.Labels)
}

func ParseMetricName(line string) (string, []Label, error) {
	name, labels, _, err := splitMetricLine(line)
	return name, labels, err
}

func ParseMetricLine(line string) (RawMetric, error) {
	name, labels, valueStr, err := splitMetricLine(line)
	if err != nil {
		return RawMetric{}, err
	}

	var value float64
	_, err = fmt.Sscanf(valueStr, "%f", &value)
	if err != nil {
		return RawMetric{}, fmt.Errorf("%w: value is not a number", ErrInvalidMetricLine)
	}

	return RawMetric{
		Name:   name,
		Labels: labels,
		Value:  value,
	}, nil
}

func splitMetricLine(line string) (string, []Label, string, error) {
	labelStartIdx := strings.Index(line, "{")
	labelEndIdx := strings.Index(line, "}")

	var labels []Label
	var name, valueStr string

	if labelStartIdx != -1 && labelEndIdx != -1 && labelEndIdx > labelStartIdx {
		name = line[:labelStartIdx]

		labels = parseLabels(line[labelStartIdx+1 : labelEndIdx])

		valueStr = line[labelEndIdx+1:]
	} else {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return "", nil, "", ErrInvalidMetricLine
		}

		name = parts[0]
		valueStr = parts[1]
	}
	return name, labels, valueStr, nil
}

func parseLabels(labelString string) []Label {
	var labels []Label

	labelPairs := strings.Split(labelString, ",")
	for _, pair := range labelPairs {
		parts := strings.Split(strings.TrimSpace(pair), "=")
		if len(parts) == 2 {
			if labels == nil {
				labels = make([]Label, 0, 1)
			}

			labels = append(labels, Label{
				Name:  parts[0],
				Value: parts[1][1 : len(parts[1])-1],
			})
		}
	}
	return labels
}

type Bin struct {
	Value float64
	Count uint64
}

type Histogram struct {
	Name   string
	Labels []Label
	Bins   []Bin
}

type Metrics struct {
	Histograms []Histogram
}

const (
	bucketSuffix = "_bucket"
	countSuffix  = "_count"
	sumSuffix    = "_sum"
)

func ParseHistogram(metrics []RawMetric) (map[string]Histogram, []RawMetric) {
	type histogramMetrics struct {
		buckets []RawMetric
		count   *RawMetric
		sum     *RawMetric
	}

	filteredMetrics := make([]RawMetric, 0, len(metrics))

	histograms := make(map[string]histogramMetrics)
	for _, m := range metrics {
		name := trimHistogramSuffix(m.Name)
		if len(name) < len(m.Name) {
			labels, _ := m.Remove("le")
			sort.Slice(labels, func(i, j int) bool {
				return labels[i].Name < labels[j].Name
			})

			mk := MetricKey{
				Name:   name,
				Labels: labels,
			}

			s := mk.String()
			h := histograms[s]

			switch {
			case strings.HasSuffix(m.Name, bucketSuffix):
				h.buckets = append(h.buckets, m)
			case strings.HasSuffix(m.Name, countSuffix):
				h.count = &m
			case strings.HasSuffix(m.Name, sumSuffix):
				h.sum = &m
			}
			histograms[s] = h
		} else {
			filteredMetrics = append(filteredMetrics, m)
		}
	}

	out := make(map[string]Histogram, len(histograms))
	for k, hist := range histograms {
		name, labels, _, err := splitMetricLine(k + " 0")
		if err != nil {
			panic("unexpected err" + err.Error())
		}

		release := func(h *histogramMetrics) {
			if h.count != nil {
				filteredMetrics = append(filteredMetrics, *h.count)
			}

			if h.sum != nil {
				filteredMetrics = append(filteredMetrics, *h.sum)
			}

			filteredMetrics = append(filteredMetrics, h.buckets...)
		}

		if len(hist.buckets) > 0 && hist.count != nil && hist.sum != nil {
			bins, err := parseHistogramBins(hist.buckets)
			if err == nil {
				out[k] = Histogram{
					Name:   name,
					Labels: labels,
					Bins:   bins,
				}
			} else {
				release(&hist)
			}
		} else {
			release(&hist)
		}
	}

	return out, filteredMetrics
}

func parseHistogramBins(bucketMetrics []RawMetric) ([]Bin, error) {
	buckets := make([]Bin, len(bucketMetrics))
	for i, b := range bucketMetrics {
		v := b.Find("le")
		if v == "" {
			return nil, fmt.Errorf("missing \"le\" tag")
		}

		binValue, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, err
		}

		buckets[i] = Bin{
			Value: binValue,
			Count: uint64(b.Value),
		}
	}
	return buckets, nil
}

func trimHistogramSuffix(s string) string {
	s = strings.TrimSuffix(s, bucketSuffix)
	s = strings.TrimSuffix(s, countSuffix)
	s = strings.TrimSuffix(s, sumSuffix)
	return s
}
