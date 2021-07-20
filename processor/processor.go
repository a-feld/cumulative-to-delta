package processor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"me.localhost/cumulative-to-delta/tracking"
)

type processor struct {
	nextConsumer consumer.Metrics
	tracker      tracking.MetricTracker
}

var _ component.MetricsProcessor = (*processor)(nil)

func (p processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *processor) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (p *processor) Shutdown(ctx context.Context) error {
	return nil
}

func (p processor) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
	rms := md.ResourceMetrics()
	md.DataPointCount()

	for i := 0; i < rms.Len(); i++ {
		rm := rms.At(i)
		ilms := rm.InstrumentationLibraryMetrics()
		for j := 0; j < ilms.Len(); j++ {
			ilm := ilms.At(j)

			ms := ilm.Metrics()
			for k := 0; k < ms.Len(); k++ {
				m := ms.At(k)
				switch m.DataType() {
				case pdata.MetricDataTypeSum:
					ms := m.Sum()
					if ms.AggregationTemporality() == pdata.AggregationTemporalityCumulative {
						ms.SetAggregationTemporality(pdata.AggregationTemporalityDelta)

						dps := ms.DataPoints()
						for d := 0; d < dps.Len(); d++ {
							dp := dps.At(d)
							tdp := tracking.DataPoint{
								Identity: tracking.MetricIdentity{
									Resource:               rm.Resource(),
									InstrumentationLibrary: ilm.InstrumentationLibrary(),
									Metric:                 m,
									LabelsMap:              dp.LabelsMap(),
								},
								Point: tracking.MetricPoint{
									ObservedTimestamp: dp.Timestamp(),
									Value:             dp.Value(),
								},
							}
							delta := p.tracker.Convert(tdp)
							dp.SetStartTimestamp(delta.StartTimestamp)
							dp.SetValue(delta.Value.(float64))
						}
					}
				}
			}
		}
	}
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func createProcessor(_ *Config, nextConsumer consumer.Metrics) (*processor, error) {
	p := &processor{
		nextConsumer: nextConsumer,
		tracker:      tracking.MetricTracker{},
	}
	return p, nil
}
