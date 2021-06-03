package cumulativetodelta

type MetricIdentity struct {
	name string
}

func ComputeMetricIdentity(m Metric) MetricIdentity {
	return MetricIdentity{name: m.Name}
}
