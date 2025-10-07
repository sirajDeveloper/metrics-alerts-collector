package repository

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain"
)

type MetricRepository interface {
	Save(metrics *domain.Metrics)
	GetMetric(metricType, metricName string) *domain.Metrics
}
