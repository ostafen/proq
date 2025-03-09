package store

import (
	"maps"
	"sort"

	"github.com/ostafen/proq/pkg/metric"
)

type RingBuffer struct {
	next   int
	values []float64
	n      uint64

	ch    chan float64
	bindN uint64
}

func (buf *RingBuffer) Add(v float64) {
	buf.values[buf.next] = v
	buf.next = (buf.next + 1) % len(buf.values)
	buf.n++

	if buf.ch != nil {
		buf.ch <- v
	}
}

type MetricID uint32

type MetricStore struct {
	numSamples int

	nextMetricID MetricID

	index   map[string]MetricID
	metrics map[MetricID]*RingBuffer

	histograms map[string]metric.Histogram
}

func NewMetricStore(numSamples int) *MetricStore {
	return &MetricStore{
		numSamples: int(numSamples),
		histograms: make(map[string]metric.Histogram),
		index:      make(map[string]MetricID),
		metrics:    make(map[MetricID]*RingBuffer),
	}
}

func (st *MetricStore) UpdateHistograms(hs map[string]metric.Histogram) {
	maps.Copy(st.histograms, hs)
}

func (st *MetricStore) Update(m *metric.RawMetric) {
	sort.Slice(m.Labels, func(i, j int) bool {
		return m.Labels[i].Name < m.Labels[j].Value
	})

	key := metric.MetricKey{
		Name:   m.Name,
		Labels: m.Labels,
	}

	buf := st.getMetric(key)
	buf.Add(m.Value)
}

func (st *MetricStore) getMetric(key metric.MetricKey) *RingBuffer {
	s := key.String()

	id, has := st.index[s]
	if has {
		return st.metrics[id]
	}

	id = st.nextMetricID
	st.index[s] = id
	st.nextMetricID++

	buf := &RingBuffer{
		values: make([]float64, st.numSamples),
	}

	st.metrics[id] = buf
	return buf
}

type Stream struct {
	st *MetricStore

	key string
	ch  chan float64
}

func (s *Stream) Close() {
	s.st.close(s.key)
}

func (st *MetricStore) Samples(key metric.MetricKey, onSample func(float64)) int {
	id, has := st.index[key.String()]
	if !has {
		return -1
	}

	buf := st.metrics[id]

	end := min(int(buf.n), len(buf.values))

	start := (buf.next - int(buf.n) + len(buf.values)) % len(buf.values)
	for i := start; i < end; i++ {
		onSample(buf.values[i%len(buf.values)])
	}
	return int(buf.n)
}

func (st *MetricStore) GetHist(mk metric.MetricKey) *metric.Histogram {
	h, ok := st.histograms[mk.String()]
	if !ok {
		panic(mk.String())
	}
	return &h
}

func (st *MetricStore) Bind(key metric.MetricKey, outChan chan float64, n int) *Stream {
	labels := key.Labels
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Name < labels[j].Value
	})

	id, has := st.index[key.String()]
	if !has {
		return nil
	}

	buf := st.metrics[id]
	buf.ch = outChan
	buf.bindN = uint64(n)

	return &Stream{
		st:  st,
		key: key.String(),
		ch:  buf.ch,
	}
}

func (st *MetricStore) close(key string) {
	id := st.index[key]
	st.metrics[id].ch = nil
}
