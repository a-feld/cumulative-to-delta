package processor

import (
	"context"

	"go.opentelemetry.io/collector/consumer/pdata"
)

type processor struct{
	cumulativeMetricChannel chan FlatMetric
}

type FlatMetric struct {
	Resource               pdata.Resource
	InstrumentationLibrary pdata.InstrumentationLibrary
	Metric                 pdata.Metric
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
						p.cumulativeMetricChannel <- FlatMetric{
							Resource:               rm.Resource(),
							InstrumentationLibrary: ilm.InstrumentationLibrary(),
							Metric:                 m,
						}
						return true
					}
				case pdata.MetricDataTypeDoubleSum:
					if m.DoubleSum().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						p.cumulativeMetricChannel <- FlatMetric{
							Resource:               rm.Resource(),
							InstrumentationLibrary: ilm.InstrumentationLibrary(),
							Metric:                 m,
						}
						return true
					}
				case pdata.MetricDataTypeIntHistogram:
					if m.IntHistogram().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						p.cumulativeMetricChannel <- FlatMetric{
							Resource:               rm.Resource(),
							InstrumentationLibrary: ilm.InstrumentationLibrary(),
							Metric:                 m,
						}
						return true
					}
				case pdata.MetricDataTypeHistogram:
					if m.Histogram().AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						p.cumulativeMetricChannel <- FlatMetric{
							Resource:               rm.Resource(),
							InstrumentationLibrary: ilm.InstrumentationLibrary(),
							Metric:                 m,
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
