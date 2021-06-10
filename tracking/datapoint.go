package tracking

type DataPoint interface {
	MetricIdentity
	MetricValue
	Metadata() MetricMetadata
}
