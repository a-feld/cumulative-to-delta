package tracking

import (
	"sync"

	"go.opentelemetry.io/collector/model/pdata"
)

type State struct {
	Identity       MetricIdentity
	LastCumulative interface{}
	LatestValue    interface{}
	Offset         interface{}
	LastObserved   pdata.Timestamp
	Mu             sync.Mutex
}

func (s *State) Lock() {
	s.Mu.Lock()
}

func (s *State) Unlock() {
	s.Mu.Unlock()
}

type DeltaValue struct {
	StartTimestamp pdata.Timestamp
	Value          interface{}
}

type MetricTracker struct {
	States sync.Map
}

func (m *MetricTracker) Convert(in DataPoint) (out DeltaValue) {
	metricId := in.Identity()
	metricPoint := in.Point()

	hashableId := metricId.AsString()
	s, _ := m.States.LoadOrStore(hashableId, &State{
		Identity:     metricId,
		Mu:           sync.Mutex{},
		LastObserved: metricPoint.Timestamp(),
	})
	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	switch metricId.Metric().DataType() {
	case pdata.MetricDataTypeSum:
		// Convert state values to float64
		offset := state.Offset.(float64)
		value := metricPoint.Value().(float64)
		latestValue := state.LatestValue.(float64)
		lastCumulative := state.LastCumulative.(float64)

		// Detect reset on a monotonic counter
		if value < latestValue {
			offset += latestValue
		}

		// Update the total cumulative count
		// Delta will be computed as currentCumulative - lastCumulative
		currentCumulative := value + offset
		out = DeltaValue{
			StartTimestamp: state.LastObserved,
			Value:          currentCumulative - lastCumulative,
		}

		// Store state values
		state.Offset = offset
		state.LatestValue = value
		state.LastCumulative = currentCumulative
		state.LastObserved = metricPoint.Timestamp()
	default:
		m.States.Delete(hashableId)
	}

	return
}
