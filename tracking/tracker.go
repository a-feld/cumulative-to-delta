package tracking

import (
	"bytes"
	"context"
	"math"
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
	t := &metricTracker{maxStale: maxStale}
	if maxStale > 0 {
		go t.sweeper(ctx, t.removeStale)
	}
	return t
}

type metricTracker struct {
	maxStale time.Duration
	states   sync.Map
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

	// NaN is used to signal "stale" metrics.
	// These are ignored for now.
	// https://github.com/open-telemetry/opentelemetry-collector/pull/3423
	if math.IsNaN(metricPoint.FloatValue) {
		return
	}

	b := identityBufferPool.Get().(*bytes.Buffer)
	b.Reset()
	metricId.Write(b)
	hashableId := b.String()
	identityBufferPool.Put(b)

	var s interface{}
	var ok bool
	if s, ok = t.states.Load(hashableId); !ok {
		s, ok = t.states.LoadOrStore(hashableId, &State{
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

func (t *metricTracker) removeStale(staleBefore pdata.Timestamp) {
	t.states.Range(func(key, value interface{}) bool {
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
			t.states.Delete(key)
		}
		return true
	})
}

func (t *metricTracker) sweeper(ctx context.Context, remove func(pdata.Timestamp)) {
	ticker := time.NewTicker(t.maxStale)
	for {
		select {
		case currentTime := <-ticker.C:
			staleBefore := pdata.TimestampFromTime(currentTime.Add(-t.maxStale))
			remove(staleBefore)
		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}
