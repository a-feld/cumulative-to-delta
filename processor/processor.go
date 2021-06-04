package processor

import (
	"context"

	"go.opentelemetry.io/collector/consumer/pdata"
)

type processor struct{}

func (p processor) ProcessMetrics(ctx context.Context, md pdata.Metrics) (pdata.Metrics, error) {
	rms := md.ResourceMetrics()

	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)
			ms := ilm.Metrics()
			ms.RemoveIf(func(m pdata.Metric) bool {
				switch m.DataType() {
				case pdata.MetricDataTypeIntSum:
					return m.IntSum().AggregationTemporality() == pdata.AggregationTemporalityCumulative
				case pdata.MetricDataTypeDoubleSum:
					return m.DoubleSum().AggregationTemporality() == pdata.AggregationTemporalityCumulative
				case pdata.MetricDataTypeIntHistogram:
					return m.IntHistogram().AggregationTemporality() == pdata.AggregationTemporalityCumulative
				case pdata.MetricDataTypeHistogram:
					return m.Histogram().AggregationTemporality() == pdata.AggregationTemporalityCumulative
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
