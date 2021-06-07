package tracking

import (
	"hash/fnv"
	"sync"
)

var metrics *sync.Map = &sync.Map{}

type metricIdentity struct {
	hash uint64
}

func (i metricIdentity) Metric() *Metric {
	v, ok := metrics.Load(i)
	if !ok {
		return nil
	}
	return v.(*Metric)
}

func ComputeMetricIdentity(m FlatMetric) metricIdentity {
	h := fnv.New64a()
	h.Write([]byte(m.Name))
	i := metricIdentity{hash: h.Sum64()}
	metrics.LoadOrStore(i, &m)
	return i
}
