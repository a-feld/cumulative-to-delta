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
	tracker      *tracking.MetricTracker
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
				baseIdentity := tracking.MetricIdentity{
					Resource:               rm.Resource(),
					InstrumentationLibrary: ilm.InstrumentationLibrary(),
					MetricDataType:         m.DataType(),
					MetricName:             m.Name(),
					MetricDescription:      m.Description(),
					MetricUnit:             m.Unit(),
				}
				switch m.DataType() {
				case pdata.MetricDataTypeIntSum:
					ms := m.IntSum()
					if ms.AggregationTemporality() != pdata.AggregationTemporalityCumulative {
						break
					}
					ms.SetAggregationTemporality(pdata.AggregationTemporalityDelta)
					baseIdentity.MetricIsMonotonic = ms.IsMonotonic()
					p.convertDataPoints(ms.DataPoints(), baseIdentity)
				case pdata.MetricDataTypeSum:

					ms := m.Sum()
					if ms.AggregationTemporality() != pdata.AggregationTemporalityCumulative {
						break
					}
					ms.SetAggregationTemporality(pdata.AggregationTemporalityDelta)
					baseIdentity.MetricIsMonotonic = ms.IsMonotonic()
					p.convertDataPoints(ms.DataPoints(), baseIdentity)
				}
			}
		}
	}
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p processor) convertDataPoints(in interface{}, baseIdentity tracking.MetricIdentity) {
	switch dps := in.(type) {
	case pdata.DoubleDataPointSlice:
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			id := baseIdentity
			id.LabelsMap = dp.LabelsMap()
			point := tracking.MetricPoint{
				ObservedTimestamp: dp.Timestamp(),
				Value:             dp.Value(),
			}
			delta := p.tracker.Convert(tracking.DataPoint{
				Identity: id,
				Point:    point,
			})
			dp.SetStartTimestamp(delta.StartTimestamp)
			dp.SetValue(delta.Value.(float64))
		}
	case pdata.IntDataPointSlice:
		for i := 0; i < dps.Len(); i++ {
			dp := dps.At(i)
			id := baseIdentity
			id.LabelsMap = dp.LabelsMap()
			point := tracking.MetricPoint{
				ObservedTimestamp: dp.Timestamp(),
				Value:             dp.Value(),
			}
			delta := p.tracker.Convert(tracking.DataPoint{
				Identity: id,
				Point:    point,
			})
			dp.SetStartTimestamp(delta.StartTimestamp)
			dp.SetValue(delta.Value.(int64))
		}
	}
}

func createProcessor(_ *Config, nextConsumer consumer.Metrics) (*processor, error) {
	p := &processor{
		nextConsumer: nextConsumer,
		tracker:      &tracking.MetricTracker{},
	}
	return p, nil
}
