package tracking

type DataPoint interface {
	Identity() MetricIdentity
	MetricValue
}
