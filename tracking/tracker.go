package tracking

import (
	"sync"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

type State struct {
	Identity          MetricIdentity
	Valid             bool
	CurrentCumulative interface{}
	LastCumulative    interface{}
	LatestValue       interface{}
	Offset            interface{}
	mu                sync.Mutex
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
}

func (m *MetricTracker) Record(in DataPoint) {
	metricId := in.Identity()
	hashableId := metricId.AsString()
	s, _ := m.States.LoadOrStore(hashableId, &State{Identity: metricId, mu: sync.Mutex{}})
	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	// Assume state values will be assigned to here, so mark the state as valid
	state.Valid = true

	// Compute updated offset if applicable
	switch metricId.Metric().DataType() {
	case pdata.MetricDataTypeSum:
		// Convert state values to float64
		offset := state.Offset.(float64)
		value := in.Value().(float64)
		latestValue := state.LatestValue.(float64)

		// Detect reset on a monotonic counter
		if value < latestValue {
			offset += latestValue
		}

		// Update the total cumulative count
		// Delta will be computed as totalCumulative - lastCumulative
		currentCumulative := value + offset

		// Store state values
		state.Offset = offset
		state.LatestValue = value
		state.CurrentCumulative = currentCumulative
	}

	// TODO: persist to disk
}

func (m *MetricTracker) Flush() pdata.Metrics {
	metrics := pdata.NewMetrics()
	t := pdata.TimestampFromTime(time.Now())

	m.States.Range(func(_, value interface{}) bool {
		state := value.(*State)
		state.Lock()
		defer state.Unlock()

		// Only attempt to process valid states (states can be added while flushing)
		if !state.Valid {
			return true
		}

		identity := state.Identity
		rms := metrics.ResourceMetrics().AppendEmpty()
		identity.Resource().CopyTo(rms.Resource())

		ilms := rms.InstrumentationLibraryMetrics().AppendEmpty()
		identity.InstrumentationLibrary().CopyTo(ilms.InstrumentationLibrary())

		ms := ilms.Metrics()
		me := ms.AppendEmpty()
		identity.Metric().CopyTo(me)

		switch me.DataType() {
		case pdata.MetricDataTypeSum:
			currentCumulative := state.CurrentCumulative.(float64)
			lastCumulative := state.LastCumulative.(float64)
			v := currentCumulative - lastCumulative
			dp := me.Sum().DataPoints().AppendEmpty()
			dp.SetStartTimestamp(m.LastFlushTime)
			dp.SetTimestamp(t)
			dp.SetValue(v)
			identity.LabelsMap().CopyTo(dp.LabelsMap())
		}

		state.LastCumulative = state.CurrentCumulative
		return true
	})

	m.LastFlushTime = t

	// TODO: flush m.States to disk via json marshal
	// Once Flush is called, any metric deltas are considered "sent"
	return metrics
}
