package cumulativetodelta

import "sync"

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

type Metric struct {
	Name  string
	Value float64
}

type MetricTracker struct {
	States sync.Map
}

func (m *MetricTracker) Record(in Metric) {
	metricId := ComputeMetricIdentity(in)
	s, _ := m.States.LoadOrStore(metricId, &State{mu: sync.Mutex{}})
	state := s.(*State)
	state.Lock()
	defer state.Unlock()

	// Compute updated offset if applicable
	if in.Value < state.LatestValue {
		state.Offset += state.LatestValue
	}
	state.LatestValue = in.Value
	state.RunningTotal = in.Value + state.Offset

	// TODO: persist to disk
}

func (m *MetricTracker) Flush() []Metric {
	var metrics []Metric

	m.States.Range(func(key, value interface{}) bool {
		identity := key.(metricIdentity)
		state := value.(*State)
		state.Lock()
		defer state.Unlock()

		metric := identity.Metric()
		metrics = append(metrics, Metric{
			Name:  metric.Name,
			Value: state.RunningTotal - state.LastFlushed,
		})
		state.LastFlushed = state.RunningTotal
		return true
	})

	// TODO: flush m.States to disk via json marshal
	// Once Flush is called, any metric deltas are considered "sent"
	return metrics
}
