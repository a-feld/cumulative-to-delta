package processor

import (
	"bytes"
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/consumer/pdata"
	tracetranslator "go.opentelemetry.io/collector/translator/trace"
	"me.localhost/cumulative-to-delta/tracking"
)

type processor struct {
	nextConsumer     consumer.Metrics
	dataPointChannel chan dataPoint
	goroutines       sync.WaitGroup
	shutdownC        chan struct{}
	tracker          tracking.MetricTracker
	flushInterval    time.Duration
}

var _ component.MetricsProcessor = (*processor)(nil)

type metricIdentity struct {
	resource               pdata.Resource
	instrumentationLibrary pdata.InstrumentationLibrary
	metric                 pdata.Metric
	labelsMap              pdata.StringMap
}

func (mi *metricIdentity) Resource() pdata.Resource {
	return mi.resource
}

func (mi *metricIdentity) InstrumentationLibrary() pdata.InstrumentationLibrary {
	return mi.instrumentationLibrary
}

func (mi *metricIdentity) Metric() pdata.Metric {
	return mi.metric
}

func (mi *metricIdentity) LabelsMap() pdata.StringMap {
	return mi.labelsMap
}

func (mi *metricIdentity) Identity() string {
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
	return h.String()
}

type metricValue struct {
	timestamp pdata.Timestamp
	value     interface{}
}

func (mv metricValue) Timestamp() pdata.Timestamp {
	return mv.timestamp
}

func (mv metricValue) Value() interface{} {
	return mv.value
}

type dataPoint struct {
	identity metricIdentity
	value    metricValue
}

func (dp dataPoint) Metadata() tracking.MetricMetadata {
	return &dp.identity
}

func (dp dataPoint) Identity() string {
	return dp.identity.Identity()
}

func (dp dataPoint) Timestamp() pdata.Timestamp {
	return dp.value.Timestamp()
}

func (dp dataPoint) Value() interface{} {
	return dp.value.Value()
}

func (p processor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *processor) Start(ctx context.Context, host component.Host) error {
	p.goroutines.Add(1)
	go p.startProcessingCycle()
	return nil
}

func (p *processor) Shutdown(ctx context.Context) error {
	close(p.shutdownC)
	p.goroutines.Wait()
	return nil
}

func (p *processor) startProcessingCycle() {
	defer p.goroutines.Done()
	flushTicker := time.NewTicker(p.flushInterval)
	for {
	DONE:
		select {
		case dp := <-p.dataPointChannel:
			p.tracker.Record(dp)
		case <-flushTicker.C:
			metrics := p.tracker.Flush()
			p.nextConsumer.ConsumeMetrics(context.Background(), metrics)
		case <-p.shutdownC:
			flushTicker.Stop()
			break DONE
		}
	}
}

func (p processor) ConsumeMetrics(ctx context.Context, md pdata.Metrics) error {
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
									value:     dp.Value(),
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
					// TODO: implement
					return true
				case pdata.MetricDataTypeHistogram:
					// TODO: implement
					return true
				}
				return false
			})
		}
	}
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

func createProcessor(cfg *Config, nextConsumer consumer.Metrics) (*processor, error) {
	p := &processor{
		nextConsumer:     nextConsumer,
		dataPointChannel: make(chan dataPoint),
		goroutines:       sync.WaitGroup{},
		shutdownC:        make(chan struct{}),
		tracker: tracking.MetricTracker{
			LastFlushTime: pdata.TimestampFromTime(time.Now()),
			States:        sync.Map{},
			Metadata:      make(map[string]tracking.MetricMetadata),
		},
		flushInterval: 60 * time.Second,
	}
	return p, nil
}
