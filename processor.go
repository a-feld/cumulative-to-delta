package cumulativetodeltaprocessor

import (
	"context"

	"github.com/a-feld/cumulativetodeltaprocessor/tracking"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

type processor struct {
	logger         *zap.Logger
	monotonicOnly  bool
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
					if p.monotonicOnly && !ms.IsMonotonic() {
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
					if p.monotonicOnly && !ms.IsMonotonic() {
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
			id.StartTimestamp = dp.StartTimestamp()
			id.LabelsMap = dp.LabelsMap()
			point := tracking.ValuePoint{
				ObservedTimestamp: dp.Timestamp(),
				FloatValue:        dp.Value(),
			}
			trackingPoint := tracking.MetricPoint{
				Identity: id,
				Point:    point,
			}
			delta, valid := p.tracker.Convert(trackingPoint)
			p.logger.Debug("cumulative-to-delta", zap.Any("id", id), zap.Any("point", point), zap.Any("delta", delta), zap.Bool("valid", valid))

			// TODO: add comment about non-monotonic cumulative metrics
			if !valid {
				return true
			}
			dp.SetStartTimestamp(delta.StartTimestamp)
			dp.SetValue(delta.FloatValue)
			return false
		})
	case pdata.IntDataPointSlice:
		dps.RemoveIf(func(dp pdata.IntDataPoint) bool {
			id := baseIdentity
			id.StartTimestamp = dp.StartTimestamp()
			id.LabelsMap = dp.LabelsMap()
			point := tracking.ValuePoint{
				ObservedTimestamp: dp.Timestamp(),
				IntValue:          dp.Value(),
			}
			trackingPoint := tracking.MetricPoint{
				Identity: id,
				Point:    point,
			}
			delta, valid := p.tracker.Convert(trackingPoint)
			p.logger.Debug("cumulative-to-delta", zap.Any("id", id), zap.Any("point", point), zap.Any("delta", delta), zap.Bool("valid", valid))
			if !valid {
				return true
			}
			dp.SetStartTimestamp(delta.StartTimestamp)
			dp.SetValue(delta.IntValue)
			return false
		})
	}
}

func createProcessor(cfg *Config, params component.ProcessorCreateSettings, nextConsumer consumer.Metrics) (*processor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	p := &processor{
		logger:        params.Logger,
		monotonicOnly: cfg.MonotonicOnly,
		nextConsumer:  nextConsumer,
		cancelFunc:    cancel,
		tracker:       tracking.NewMetricTracker(ctx, cfg.MaxStale),
	}
	if len(cfg.Metrics) > 0 {
		p.enabledMetrics = make(map[string]struct{}, len(cfg.Metrics))
		for _, m := range cfg.Metrics {
			p.enabledMetrics[m] = struct{}{}
		}
	}
	return p, nil
}
