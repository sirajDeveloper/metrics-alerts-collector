package domain

type MetricType string

const (
	Gauge   MetricType = "Gauge"
	Counter MetricType = "Counter"
)

type Metric struct {
	Name  string
	Type  MetricType
	Value any
}
