package cumulativetodeltaprocessor

import (
	"context"

	"github.com/a-feld/cumulativetodeltaprocessor/tracking"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
)

type processor struct {
	nextConsumer   consumer.Metrics
	cancelFunc     context.CancelFunc
	tracker        tracking.MetricTracker
	enabledMetrics map[string]struct{}
}

var _ component.MetricsProcessor = (*processor)(nil)

func (p processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *processor) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (p *processor) Shutdown(ctx context.Context) error {
	p.cancelFunc()
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
				if p.enabledMetrics != nil {
					if _, ok := p.enabledMetrics[m.Name()]; !ok {
						continue
					}
				}
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
					baseIdentity.MetricIsMonotonic = ms.IsMonotonic()
					p.convertDataPoints(ms.DataPoints(), baseIdentity)
					ms.SetAggregationTemporality(pdata.AggregationTemporalityDelta)
				case pdata.MetricDataTypeSum:
					ms := m.Sum()
					if ms.AggregationTemporality() != pdata.AggregationTemporalityCumulative {
						break
					}
					baseIdentity.MetricIsMonotonic = ms.IsMonotonic()
					p.convertDataPoints(ms.DataPoints(), baseIdentity)
					ms.SetAggregationTemporality(pdata.AggregationTemporalityDelta)
				}
			}
		}
	}
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func (p processor) convertDataPoints(in interface{}, baseIdentity tracking.MetricIdentity) {
	switch dps := in.(type) {
	case pdata.DoubleDataPointSlice:
		dps.RemoveIf(func(dp pdata.DoubleDataPoint) bool {
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

			// TODO: add comment about non-monotonic cumulative metrics
			if delta.Value == nil {
				return true
			}
			dp.SetStartTimestamp(delta.StartTimestamp)
			dp.SetValue(delta.Value.(float64))
			return false
		})
	case pdata.IntDataPointSlice:
		dps.RemoveIf(func(dp pdata.IntDataPoint) bool {
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
			if delta.Value == nil {
				return true
			}
			dp.SetStartTimestamp(delta.StartTimestamp)
			dp.SetValue(delta.Value.(int64))
			return false
		})
	}
}

func createProcessor(cfg *Config, nextConsumer consumer.Metrics) (*processor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &processor{
		nextConsumer: nextConsumer,
		cancelFunc:   cancel,
		tracker:      tracking.NewMetricTracker(ctx, cfg.MaxStale),
	}
	if len(cfg.Metrics) > 0 {
		p.enabledMetrics = make(map[string]struct{})
		for _, m := range cfg.Metrics {
			p.enabledMetrics[m] = struct{}{}
		}
	}
	return p, nil
}
