package repository

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type MetricRepository interface {
	Save(metrics *model.Metrics)
	GetMetric(metricType, metricName string) *model.Metrics
	GetAll() []*model.Metrics
}
