package tracking

import (
	"sync"

	"go.opentelemetry.io/collector/model/pdata"
)

type State struct {
	Identity    MetricIdentity
	LatestPoint MetricPoint
	mu          sync.Mutex
}

func (s *State) Lock() {
	s.mu.Lock()
}

func (s *State) Unlock() {
	s.mu.Unlock()
}

type DeltaValue struct {
	StartTimestamp pdata.Timestamp
	Value          interface{}
}

type MetricTracker struct {
	States sync.Map
}

func (m *MetricTracker) Convert(in DataPoint) (out DeltaValue) {
	metricId := in.Identity
	switch metricId.MetricDataType {
	case pdata.MetricDataTypeSum:
	case pdata.MetricDataTypeIntSum:
	default:
		return
	}
	metricPoint := in.Point

	hashableId := metricId.AsString()
	s, ok := m.States.LoadOrStore(hashableId, &State{
		Identity:    metricId,
		mu:          sync.Mutex{},
		LatestPoint: metricPoint,
	})

	if !ok {
		if metricId.MetricIsMonotonic {
			out = DeltaValue{
				StartTimestamp: metricPoint.ObservedTimestamp,
				Value:          metricPoint.Value,
			}
		}
		return
	}

	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	out.StartTimestamp = state.LatestPoint.ObservedTimestamp

	switch metricId.MetricDataType {
	case pdata.MetricDataTypeSum:
		// Convert state values to float64
		value := metricPoint.Value.(float64)
		latestValue := state.LatestPoint.Value.(float64)
		delta := value - latestValue

		// Detect reset on a monotonic counter
		if metricId.MetricIsMonotonic && value < latestValue {
			delta = value
		}

		out.Value = delta
	case pdata.MetricDataTypeIntSum:
		// Convert state values to int64
		value := metricPoint.Value.(int64)
		latestValue := state.LatestPoint.Value.(int64)
		delta := value - latestValue

		// Detect reset on a monotonic counter
		if metricId.MetricIsMonotonic && value < latestValue {
			delta = value
		}

		out.Value = delta
	}

	state.LatestPoint = metricPoint
	return
}
