package tracking

import (
	"go.opentelemetry.io/collector/model/pdata"
)

type MetricPoint struct {
	ObservedTimestamp pdata.Timestamp
	Value             interface{}
}

type HistogramValue interface {
	Count() uint64
	Sum() interface{} // can be float or int
	BucketCounts() []uint64
	ExplicitBounds() []float64
}
