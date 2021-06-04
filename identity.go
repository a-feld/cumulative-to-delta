package cumulativetodelta

import (
	"hash/fnv"
)

type MetricIdentity struct {
	hash uint64
}

func ComputeMetricIdentity(m Metric) MetricIdentity {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	return MetricIdentity{hash: h.Sum64()}
}
