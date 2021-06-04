package cumulativetodelta

import (
	"hash/fnv"
	"sync/atomic"
)

var metrics map[MetricIdentity]atomic.Value

type MetricIdentity struct {
	hash uint64
}

func (i MetricIdentity) Name() string {
	v := metrics[i]
	return v.Load().(Metric).Name
}

func ComputeMetricIdentity(m Metric) MetricIdentity {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	i := MetricIdentity{hash: h.Sum64()}
	v := metrics[i]
	v.Store(m)
	return i
}
