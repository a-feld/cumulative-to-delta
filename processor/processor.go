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
			for k := 0; k < ms.Len(); k++ {
				m := ms.At(k)
				switch m.DataType() {
				case pdata.MetricDataTypeIntSum:
					m.IntSum().AggregationTemporality()
				case pdata.MetricDataTypeDoubleSum:
					m.DoubleSum().DataPoints().Len()
				case pdata.MetricDataTypeIntHistogram:
					m.IntHistogram().DataPoints().Len()
				case pdata.MetricDataTypeHistogram:
					m.Histogram().DataPoints().Len()
				default:
					m.Summary().DataPoints().Len()
				}
			}
		}
	}
	return md, nil
}

func createProcessor(cfg *Config) (*processor, error) {
	return nil, nil
}
