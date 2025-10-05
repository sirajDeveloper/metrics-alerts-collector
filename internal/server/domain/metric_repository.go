package domain

type MetricRepository interface {
	Put(metrics *Metrics)
}
