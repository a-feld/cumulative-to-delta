package processor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "delta_converter"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

// NewFactory returns a new factory for the Resource processor.
func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		processorhelper.WithMetrics(createMetricsProcessor),
	)
}

func createDefaultConfig() config.Processor {
	return &Config{}
}

func createMetricsProcessor(
	_ context.Context,
	_ component.ProcessorCreateParams,
	cfg config.Processor,
	nextConsumer consumer.Metrics) (component.MetricsProcessor, error) {
	processor, err := createProcessor(cfg.(*Config))
	if err != nil {
		return nil, err
	}
	return processorhelper.NewMetricsProcessor(
		cfg,
		nextConsumer,
		processor,
		processorhelper.WithCapabilities(processorCapabilities))
}
