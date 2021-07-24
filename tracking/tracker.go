package tracking

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

type State struct {
	Identity  MetricIdentity
	PrevPoint MetricPoint
	mu        sync.Mutex
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
	var s interface{}
	var ok bool
	if s, ok = t.States.Load(hashableId); !ok {
		s, ok = t.States.LoadOrStore(hashableId, &State{
			Identity:  metricId,
			PrevPoint: metricPoint,
		})
	}

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

	out.StartTimestamp = state.PrevPoint.ObservedTimestamp

	switch metricId.MetricDataType {
	case pdata.MetricDataTypeSum:
		// Convert state values to float64
		value := metricPoint.Value.(float64)
		prevValue := state.PrevPoint.Value.(float64)
		delta := value - prevValue

		// Detect reset on a monotonic counter
		if metricId.MetricIsMonotonic && value < prevValue {
			delta = value
		}

		out.Value = delta
	case pdata.MetricDataTypeIntSum:
		// Convert state values to int64
		value := metricPoint.Value.(int64)
		prevValue := state.PrevPoint.Value.(int64)
		delta := value - prevValue

		// Detect reset on a monotonic counter
		if metricId.MetricIsMonotonic && value < prevValue {
			delta = value
		}

		out.Value = delta
	}

	state.PrevPoint = metricPoint
	return
}

func (t *metricTracker) RemoveStale(staleBefore pdata.Timestamp) {
	t.States.Range(func(key, value interface{}) bool {
		s := value.(*State)

		// There is a known race condition here.
		// Because the state may be in the process of updating at the
		// same time as the stale removal, there is a chance that we
		// will remove a "stale" state that is in the process of
		// updating. This can only happen when datapoints arrive around
		// the expiration time.
		//
		// In this case, the possible outcomes are:
		//	* Updating goroutine wins, point will not be stale
		//	* Stale removal wins, updating goroutine will still see
		//	  the removed state but the state after the update will
		//	  not be persisted. The next update will load an entirely
		//	  new state.
		s.Lock()
		lastObserved := s.PrevPoint.ObservedTimestamp
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

	go func() {
		ticker := time.NewTicker(t.MaxStale)
		for {
			select {
			case currentTime := <-ticker.C:
				staleBefore := pdata.TimestampFromTime(currentTime.Add(-t.MaxStale))
				t.RemoveStale(staleBefore)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}
