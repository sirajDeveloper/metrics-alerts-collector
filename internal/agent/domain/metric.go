package domain

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MetricType string

type Metric struct {
	Name  string
	Type  MetricType
	Value any
}
