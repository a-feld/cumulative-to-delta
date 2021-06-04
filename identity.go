package cumulativetodelta

import (
	"hash/fnv"
	"sync"
)

var metrics sync.Map = sync.Map{}

type metricIdentity struct {
	hash uint64
}

func (i metricIdentity) Metric() *Metric {
	v, _ := metrics.Load(i)
	// This will panic if v is nil, which is expected behavior
	// All calls to metrics.Load should succeed
	return v.(*Metric)
}

func ComputeMetricIdentity(m Metric) metricIdentity {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	i := metricIdentity{hash: h.Sum64()}
	metrics.LoadOrStore(i, &m)
	return i
}
