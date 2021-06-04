package processor

import (
	"context"

	"go.opentelemetry.io/collector/consumer/pdata"
)

type processor struct{}

func (p processor) ProcessMetrics(ctx context.Context, m pdata.Metrics) (pdata.Metrics, error) {
	return m, nil
}

func createProcessor(cfg *Config) (*processor, error) {
	return nil, nil
}
