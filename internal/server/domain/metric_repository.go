package domain

type MetricRepository interface {
	Save(metrics *Metrics)
	GetMetric(metricType, metricName string) *Metrics
}
