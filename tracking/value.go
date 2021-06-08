package tracking

import (
	"go.opentelemetry.io/collector/consumer/pdata"
)

type MetricValue interface {
	Timestamp() pdata.Timestamp
	Value() float64
}
