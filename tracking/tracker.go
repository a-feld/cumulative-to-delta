package tracking

import (
	"bytes"
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

// Allocate a minimum of 64 bytes to the builder initially
const initialBytes = 64

var identityBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, initialBytes))
	},
}

type State struct {
	Identity  MetricIdentity
	PrevPoint ValuePoint
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
	FloatValue     float64
	IntValue       int64
}

type MetricTracker interface {
	Convert(MetricPoint) (DeltaValue, bool)
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

func (t *metricTracker) Convert(in MetricPoint) (out DeltaValue, valid bool) {
	metricId := in.Identity
	switch metricId.MetricDataType {
	case pdata.MetricDataTypeSum:
	case pdata.MetricDataTypeIntSum:
	default:
		return
	}

	metricPoint := in.Point

	b := identityBufferPool.Get().(*bytes.Buffer)
	b.Reset()
	metricId.Write(b)
	hashableId := b.String()
	identityBufferPool.Put(b)

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
				FloatValue:     metricPoint.FloatValue,
				IntValue:       metricPoint.IntValue,
			}
			valid = true
		}
		return
	}
	valid = true

	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	out.StartTimestamp = state.PrevPoint.ObservedTimestamp

	switch metricId.MetricDataType {
	case pdata.MetricDataTypeSum:
		// Convert state values to float64
		value := metricPoint.FloatValue
		prevValue := state.PrevPoint.FloatValue
		delta := value - prevValue

		// Detect reset on a monotonic counter
		if metricId.MetricIsMonotonic && value < prevValue {
			delta = value
		}

		out.FloatValue = delta
	case pdata.MetricDataTypeIntSum:
		// Convert state values to int64
		value := metricPoint.IntValue
		prevValue := state.PrevPoint.IntValue
		delta := value - prevValue

		// Detect reset on a monotonic counter
		if metricId.MetricIsMonotonic && value < prevValue {
			delta = value
		}

		out.IntValue = delta
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
