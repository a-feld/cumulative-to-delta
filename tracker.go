package cumulativetodelta

import "sync"

type MetricIdentity struct {
	name string
}

type State struct {
	CurrentTotal     float64
	CurrentValue     float64
	Offset           float64
	LastFlushedValue float64
}
type Metric struct {
	Name  string
	Value float64
}

func (m Metric) Identity() MetricIdentity {
	return MetricIdentity{name: m.Name}
}

type MetricTracker struct {
	mu     sync.Mutex
	States map[MetricIdentity]State
}

func (m *MetricTracker) Record(in Metric) {
	var total, lastFlushed, offset float64
	m.mu.Lock()
	defer m.mu.Unlock()
	if state, ok := m.States[in.Identity()]; ok {
		if in.Value < state.CurrentValue {
			offset = state.Offset + state.CurrentValue
		} else {
			offset = state.Offset
		}
		lastFlushed = state.LastFlushedValue
	}
	total = in.Value + offset
	m.States[in.Identity()] = State{
		CurrentTotal:     total,
		CurrentValue:     in.Value,
		LastFlushedValue: lastFlushed,
		Offset:           offset,
	}
}

func (m *MetricTracker) Flush() []Metric {
	m.mu.Lock()
	defer m.mu.Unlock()
	metrics := make([]Metric, len(m.States), 0)
	for identity, state := range m.States {
		metrics = append(metrics, Metric{
			Name:  identity.name,
			Value: state.CurrentTotal - state.LastFlushedValue,
		})
		state.LastFlushedValue = state.CurrentTotal
	}
	return metrics
}
