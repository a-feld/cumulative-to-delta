package tracking

import (
	"hash/fnv"
	"sync"

	"go.opentelemetry.io/collector/consumer/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
	"localhost.me/cumulative-to-delta/processor"
)

var metrics *sync.Map = &sync.Map{}

type metricIdentity struct {
	hash uint64
}

func (i metricIdentity) Metric() *processor.FlatMetric {
	v, ok := metrics.Load(i)
	if !ok {
		return nil
	}
	return v.(*processor.FlatMetric)
}

func ComputeMetricIdentity(mi processor.FlatMetric) metricIdentity {
	h := fnv.New64a()
	h.Write([]byte("r;"))
	mi.Resource.Attributes().Sort().Range(func(k string, v pdata.AttributeValue) bool {
		h.Write([]byte(k))
		h.Write([]byte(";"))
		h.Write([]byte(tracetranslator.AttributeValueToString(v)))
		h.Write([]byte(";"))
		return true
	})

	h.Write([]byte(";i;"))
	h.Write([]byte(mi.InstrumentationLibrary.Name()))
	h.Write([]byte(";"))
	h.Write([]byte(mi.InstrumentationLibrary.Version()))

	h.Write([]byte(";m;"))
	h.Write([]byte(mi.Metric.Name()))
	h.Write([]byte(";"))
	h.Write([]byte(mi.Metric.Description()))
	h.Write([]byte(";"))
	h.Write([]byte(mi.Metric.Unit()))

	h.Write([]byte(";d;"))
	m := mi.Metric
	switch m.DataType() {
	case pdata.MetricDataTypeIntSum:
		dp := mi.DataPoint.(*pdata.IntDataPoint)
		dp.LabelsMap().Sort().Range(func(k, v string) bool {
			h.Write([]byte(k))
			h.Write([]byte(";"))
			h.Write([]byte(v))
			h.Write([]byte(";"))
			return true
		})
	case pdata.MetricDataTypeDoubleSum:
		dp := mi.DataPoint.(*pdata.DoubleDataPoint)
		dp.LabelsMap().Sort().Range(func(k, v string) bool {
			h.Write([]byte(k))
			h.Write([]byte(";"))
			h.Write([]byte(v))
			h.Write([]byte(";"))
			return true
		})
	case pdata.MetricDataTypeIntHistogram:
		dp := mi.DataPoint.(*pdata.IntHistogramDataPoint)
		dp.LabelsMap().Sort().Range(func(k, v string) bool {
			h.Write([]byte(k))
			h.Write([]byte(";"))
			h.Write([]byte(v))
			h.Write([]byte(";"))
			return true
		})
	case pdata.MetricDataTypeHistogram:
		dp := mi.DataPoint.(*pdata.HistogramDataPoint)
		dp.LabelsMap().Sort().Range(func(k, v string) bool {
			h.Write([]byte(k))
			h.Write([]byte(";"))
			h.Write([]byte(v))
			h.Write([]byte(";"))
			return true
		})
	}
	i := metricIdentity{hash: h.Sum64()}
	metrics.LoadOrStore(i, &mi)
	return i
}
