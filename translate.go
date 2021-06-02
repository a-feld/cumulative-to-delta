package cumulativetodelta

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

type MetricAggregator struct {
	States map[MetricIdentity]State
}

func (m MetricAggregator) Record(in Metric) {
	// FIXME: this is not thread safe
	lastValue := 0.0

	if currentState, ok := m.States[in.Identity()]; ok {
		lastValue = currentState.LastFlushedValue
	}
	m.States[in.Identity()] = State{
		CurrentValue:     in.Value,
		LastFlushedValue: lastValue,
	}
}

func (m MetricAggregator) Flush() []Metric {
	// FIXME: this is not thread safe
	metrics := make([]Metric, len(m.States), 0)
	for identity, state := range m.States {
		metrics = append(metrics, Metric{
			Name:  identity.name,
			Value: state.CurrentValue - state.LastFlushedValue,
		})
	}
	return metrics
}
