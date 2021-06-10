package tracking

import "go.opentelemetry.io/collector/consumer/pdata"

type MetricIdentity interface {
	Identity() string
}

type MetricMetadata interface {
	Resource() pdata.Resource
	InstrumentationLibrary() pdata.InstrumentationLibrary
	Metric() pdata.Metric
	LabelsMap() pdata.StringMap
}
