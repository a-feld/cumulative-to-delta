package tracking

type DataPoint interface {
	Identity() MetricIdentity
	Value() MetricValue
}
