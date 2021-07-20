package tracking

type DataPoint interface {
	Identity() MetricIdentity
	Point() MetricPoint
}
