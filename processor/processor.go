package processor

import (
	"bytes"
	"context"

	"go.opentelemetry.io/collector/consumer/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
)

type processor struct {
	dataPointChannel chan dataPoint
}

type metricIdentity struct {
	resource               pdata.Resource
	instrumentationLibrary pdata.InstrumentationLibrary
	metric                 pdata.Metric
	labelsMap              pdata.StringMap
}

func (mi *metricIdentity) Identity() []byte {
	h := bytes.Buffer{}
	h.Write([]byte("r;"))
	mi.resource.Attributes().Sort().Range(func(k string, v pdata.AttributeValue) bool {
		h.Write([]byte(k))
		h.Write([]byte(";"))
		h.Write([]byte(tracetranslator.AttributeValueToString(v)))
		h.Write([]byte(";"))
		return true
	})

	h.Write([]byte(";i;"))
	h.Write([]byte(mi.instrumentationLibrary.Name()))
	h.Write([]byte(";"))
	h.Write([]byte(mi.instrumentationLibrary.Version()))

	h.Write([]byte(";m;"))
	h.Write([]byte(mi.metric.Name()))
	h.Write([]byte(";"))
	h.Write([]byte(mi.metric.Description()))
	h.Write([]byte(";"))
	h.Write([]byte(mi.metric.Unit()))

	h.Write([]byte(";l;"))
	mi.labelsMap.Sort().Range(func(k, v string) bool {
		h.Write([]byte(k))
		h.Write([]byte(";"))
		h.Write([]byte(v))
		h.Write([]byte(";"))
		return true
	})
	return h.Bytes()
}

type metricValue struct {
	timestamp pdata.Timestamp
	value     float64
}

func (mv metricValue) Timestamp() pdata.Timestamp {
	return mv.timestamp
}

func (mv metricValue) Value() float64 {
	return mv.value
}

type dataPoint struct {
	identity metricIdentity
	value    metricValue
}

func (p processor) ProcessMetrics(ctx context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	rms := md.ResourceMetrics()

	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			ms := ilm.Metrics()

			// Create storage for cumulative metrics

			ms.RemoveIf(func(m pdata.Metric) bool {
				switch m.DataType() {
				case pdata.MetricDataTypeIntSum:
					if m.IntSum().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						for k := 0; k < m.IntSum().DataPoints().Len(); k++ {
							dp := m.IntSum().DataPoints().At(k)
							p.dataPointChannel <- dataPoint{
								identity: metricIdentity{
									resource:               rm.Resource(),
									instrumentationLibrary: ilm.InstrumentationLibrary(),
									metric:                 m,
									labelsMap:              dp.LabelsMap(),
								},
								value: metricValue{
									timestamp: dp.Timestamp(),
									value:     float64(dp.Value()),
								},
							}
						}
						return true
					}
				case pdata.MetricDataTypeDoubleSum:
					if m.DoubleSum().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						for k := 0; k < m.DoubleSum().DataPoints().Len(); k++ {
							dp := m.DoubleSum().DataPoints().At(k)
							p.dataPointChannel <- dataPoint{
								identity: metricIdentity{
									resource:               rm.Resource(),
									instrumentationLibrary: ilm.InstrumentationLibrary(),
									metric:                 m,
									labelsMap:              dp.LabelsMap(),
								},
								value: metricValue{
									timestamp: dp.Timestamp(),
									value:     dp.Value(),
								},
							}
						}
						return true
					}
				case pdata.MetricDataTypeIntHistogram:
					if m.IntHistogram().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						for k := 0; k < m.IntHistogram().DataPoints().Len(); k++ {
							dp := m.IntHistogram().DataPoints().At(k)
							p.dataPointChannel <- dataPoint{
								identity: metricIdentity{
									resource:               rm.Resource(),
									instrumentationLibrary: ilm.InstrumentationLibrary(),
									metric:                 m,
									labelsMap:              dp.LabelsMap(),
								},
								value: metricValue{
									timestamp: dp.Timestamp(),
									value:     0, // FIXME
								},
							}
						}
						return true
					}
				case pdata.MetricDataTypeHistogram:
					if m.Histogram().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						for k := 0; k < m.Histogram().DataPoints().Len(); k++ {
							dp := m.Histogram().DataPoints().At(k)
							p.dataPointChannel <- dataPoint{
								identity: metricIdentity{
									resource:               rm.Resource(),
									instrumentationLibrary: ilm.InstrumentationLibrary(),
									metric:                 m,
									labelsMap:              dp.LabelsMap(),
								},
								value: metricValue{
									timestamp: dp.Timestamp(),
									value:     0, // FIXME
								},
							}
						}
						return true
					}
				}
				return false
			})
		}
	}
	return md, nil
}

func createProcessor(cfg *Config) (*processor, error) {
	return nil, nil
}
