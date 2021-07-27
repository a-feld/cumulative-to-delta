package cumulativetodeltaprocessor

import (
	"time"

	"go.opentelemetry.io/collector/config"
)

// Config defines configuration for Resource processor.
type Config struct {
	config.ProcessorSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct

	// The total time a state entry will live past the time it was last seen. Set to 0 to retain state indefinitely.
	MaxStale time.Duration `mapstructure:"max_stale"`

	// List of cumulative metrics to convert to delta. Default: converts all cumulative metrics to delta.
	Metrics []string `mapstructure:"metrics"`

	// Set to true in order to convert only monotonic metrics
	MonotonicOnly bool `mapstructure:"monotonic_only"`
}

var _ config.Processor = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	return nil
}
