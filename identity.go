package cumulativetodelta

import (
	"hash/fnv"
	"sync"
)

var metrics sync.Map = sync.Map{}

type MetricIdentity struct {
	hash uint64
}

func (i MetricIdentity) Metric() *Metric {
	v, ok := metrics.Load(i)
	if !ok {
		panic("Identity exists without an entry containing the original metric")
	}
	return v.(*Metric)
}

func ComputeMetricIdentity(m Metric) MetricIdentity {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	i := MetricIdentity{hash: h.Sum64()}
	metrics.LoadOrStore(i, &m)
	return i
}
