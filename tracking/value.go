package tracking

import (
	"go.opentelemetry.io/collector/model/pdata"
)

type MetricPoint struct {
	ObservedTimestamp pdata.Timestamp
	FloatValue        float64
	IntValue          int64
}

type HistogramValue interface {
	Count() uint64
	Sum() interface{} // can be float or int
	BucketCounts() []uint64
	ExplicitBounds() []float64
}
