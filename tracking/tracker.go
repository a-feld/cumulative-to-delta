package tracking

import (
	"sync"
	"time"

	"go.opentelemetry.io/collector/consumer/pdata"
)

type State struct {
	RunningTotal float64
	LatestValue  float64
	Offset       float64
	LastFlushed  float64
	mu           sync.Mutex
}

func (s *State) Lock() {
	s.mu.Lock()
}

func (s *State) Unlock() {
	s.mu.Unlock()
}

type MetricTracker struct {
	LastFlushTime pdata.Timestamp
	States        sync.Map
	Metadata      map[string]MetricMetadata
}

func (m *MetricTracker) Record(in DataPoint) {
	metricId := in.Identity()
	s, ok := m.States.LoadOrStore(metricId, &State{mu: sync.Mutex{}})
	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	if !ok {
		m.Metadata[metricId] = in.Metadata()
	}

	// Compute updated offset if applicable
	value := in.Value().(float64)
	if value < state.LatestValue {
		state.Offset += state.LatestValue
	}
	state.LatestValue = value
	state.RunningTotal = value + state.Offset

	// TODO: persist to disk
}

func (m *MetricTracker) Flush() pdata.ResourceMetricsSlice {
	metrics := pdata.NewResourceMetricsSlice()
	t := pdata.TimestampFromTime(time.Now())

	m.States.Range(func(key, value interface{}) bool {
		identity := key.(string)
		state := value.(*State)
		state.Lock()
		defer state.Unlock()

		metadata := m.Metadata[identity]
		rms := metrics.AppendEmpty()
		metadata.Resource().CopyTo(rms.Resource())

		ilms := rms.InstrumentationLibraryMetrics().AppendEmpty()
		metadata.InstrumentationLibrary().CopyTo(ilms.InstrumentationLibrary())

		ms := ilms.Metrics()
		me := metadata.Metric()
		ms.Append(me)

		v := state.RunningTotal - state.LastFlushed

		dp := me.DoubleSum().DataPoints().AppendEmpty()
		dp.SetStartTimestamp(m.LastFlushTime)
		dp.SetTimestamp(t)
		dp.SetValue(v)
		metadata.LabelsMap().CopyTo(dp.LabelsMap())

		state.LastFlushed = state.RunningTotal
		return true
	})

	m.LastFlushTime = t

	// TODO: flush m.States to disk via json marshal
	// Once Flush is called, any metric deltas are considered "sent"
	return metrics
}
