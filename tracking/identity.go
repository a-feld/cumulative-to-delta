package tracking

import "go.opentelemetry.io/collector/model/pdata"

type MetricIdentity interface {
	Resource() pdata.Resource
	InstrumentationLibrary() pdata.InstrumentationLibrary
	Metric() pdata.Metric
	LabelsMap() pdata.StringMap
	AsString() string
}
