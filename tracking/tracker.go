package tracking

import (
	"sync"

	"go.opentelemetry.io/collector/model/pdata"
)

type State struct {
	Identity    MetricIdentity
	LatestPoint MetricPoint
	Mu          sync.Mutex
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
		Identity:    metricId,
		Mu:          sync.Mutex{},
		LatestPoint: metricPoint,
	})
	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	switch metricId.Metric().DataType() {
	case pdata.MetricDataTypeSum:
		// Convert state values to float64
		value := metricPoint.Value().(float64)
		latestValue := state.LatestPoint.Value().(float64)

		// Detect reset on a monotonic counter
		offset := 0.0
		if value < latestValue {
			offset = latestValue
		}

		// Update the total cumulative count
		// Delta will be computed as currentCumulative - lastCumulative
		currentCumulative := value + offset
		out = DeltaValue{
			StartTimestamp: state.LatestPoint.ObservedTimestamp(),
			Value:          currentCumulative - latestValue,
		}

		// Store state values
		state.LatestPoint = metricPoint
	default:
		m.States.Delete(hashableId)
	}

	return
}
