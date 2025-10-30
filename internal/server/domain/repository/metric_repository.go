package repository

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type MetricRepository interface {
	Save(metrics *model.Metrics)
	GetAll() []*model.Metrics
	GetMetric(mType, name string) *model.Metrics
}

type MetricFileStorage interface {
	SaveAll(metrics []*model.Metrics)
	Save(metric *model.Metrics)
}
