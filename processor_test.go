package cumulativetodeltaprocessor

import (
	"context"
	"testing"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/model/pdata"
	"go.uber.org/zap"
)

func BenchmarkConsumeMetrics(b *testing.B) {
	c := consumertest.NewNop()
	params := component.ProcessorCreateSettings{
		Logger:         zap.NewNop(),
	}
	cfg := createDefaultConfig()
	p, err := createProcessor(cfg.(*Config), params, c)
	if err != nil {
		panic(err)
	}

	metrics := pdata.NewMetrics()
	rms := metrics.ResourceMetrics().AppendEmpty()
	r := rms.Resource()
	r.Attributes().Insert("resource", pdata.NewAttributeValueBool(true))
	ilms := rms.InstrumentationLibraryMetrics().AppendEmpty()
	ilms.InstrumentationLibrary().SetName("test")
	ilms.InstrumentationLibrary().SetVersion("0.1")
	m := ilms.Metrics().AppendEmpty()
	m.SetDataType(pdata.MetricDataTypeSum)
	m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
	m.Sum().SetIsMonotonic(true)
	dp := m.Sum().DataPoints().AppendEmpty()
	dp.LabelsMap().Insert("tag", "value")
	dp.SetValue(100.0)

	reset := func() {
		m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityCumulative)
		dp.SetValue(100.0)
	}

	// Load initial value
	p.ConsumeMetrics(context.Background(), metrics)
	reset()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p.ConsumeMetrics(context.Background(), metrics)
		reset()
	}
}
