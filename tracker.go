package cumulativetodelta

import "sync"

type MetricIdentity struct {
	name string
}

type State struct {
	CurrentValue     float64
	LowestValue		 float64
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
	lowestValue := value
	lastFlushed := 0.0
	m.mu.Lock()
	defer m.mu.Unlock()
	var delta float64
	if state, ok := m.States[in.Identity()]; ok {
		if value < state.LowestValue {
			delta = value
			lowestValue = value
		} else {
			delta = value - state.LowestValue
			lowestValue = state.LowestValue
		}
		value = state.CurrentValue + delta
		lastFlushed = state.LastFlushedValue
	}
	m.States[in.Identity()] = State{
		CurrentValue:     value,
		LastFlushedValue: lastFlushed,
		LowestValue: lowestValue,
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
