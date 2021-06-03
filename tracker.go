package cumulativetodelta

import "sync"

type MetricIdentity struct {
	name string
}

type State struct {
	CurrentValue     float64
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
	mu            sync.Mutex
	States        map[MetricIdentity]State
}

func (m *MetricTracker) Record(in Metric) {
	value := in.Value
	lastFlushed := 0.0
	m.mu.Lock()
	defer m.mu.Unlock()
	if state, ok := m.States[in.Identity()]; ok {
		if value < state.CurrentValue {
			value += state.CurrentValue
		}
		lastFlushed = state.LastFlushedValue
	}
	m.States[in.Identity()] = State{
		CurrentValue:     value,
		LastFlushedValue: lastFlushed,
	}
}

func (m *MetricTracker) Flush() []Metric {
	m.mu.Lock()
	defer m.mu.Unlock()
	metrics := make([]Metric, len(m.States), 0)
	for identity, state := range m.States {
		metrics = append(metrics, Metric{
			Name:  identity.name,
			Value: state.CurrentValue - state.LastFlushedValue,
		})
		state.LastFlushedValue = state.CurrentValue
	}
	return metrics
}
