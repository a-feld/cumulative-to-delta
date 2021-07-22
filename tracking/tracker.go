package tracking

import (
	"context"
	"sync"
	"time"

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

type MetricTracker interface {
	Convert(DataPoint) DeltaValue
}

func NewMetricTracker(ctx context.Context, maxStale time.Duration) MetricTracker {
	t := &metricTracker{MaxStale: maxStale}
	t.Start(ctx)
	return t
}

type metricTracker struct {
	MaxStale time.Duration
	States   sync.Map
}

func (t *metricTracker) Convert(in DataPoint) (out DeltaValue) {
	metricId := in.Identity
	switch metricId.MetricDataType {
	case pdata.MetricDataTypeSum:
	case pdata.MetricDataTypeIntSum:
	default:
		return
	}
	metricPoint := in.Point

	hashableId := metricId.AsString()
	s, ok := t.States.LoadOrStore(hashableId, &State{
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

func (t *metricTracker) RemoveStale(staleBefore pdata.Timestamp) {
	t.States.Range(func(key, value interface{}) bool {
		s := value.(*State)
		s.Lock()
		lastObserved := s.LatestPoint.ObservedTimestamp
		s.Unlock()
		if lastObserved < staleBefore {
			t.States.Delete(key)
		}
		return true
	})
}

func (t *metricTracker) Start(ctx context.Context) {
	if t.MaxStale == 0 {
		return
	}

	ticker := time.NewTicker(t.MaxStale)
	go func() {
		select {
		case currentTime := <-ticker.C:
			staleBefore := pdata.TimestampFromTime(currentTime.Add(-t.MaxStale))
			t.RemoveStale(staleBefore)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}()
}
