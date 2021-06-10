package tracking

import (
	"go.opentelemetry.io/collector/consumer/pdata"
)

type MetricValue interface {
	Timestamp() pdata.Timestamp
	Value() interface{}
}

type HistogramValue interface {
	Count() uint64
	Sum() interface{} // can be float or int
	BucketCounts() []uint64
	ExplicitBounds() []float64
}
