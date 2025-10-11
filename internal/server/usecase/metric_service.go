package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type MetricUpdater interface {
	MetricUpdate(metricType string, metricName string, metricValue string) error
}

type MetricService struct {
	repo repository.MetricRepository
}

func NewMetricService(repo repository.MetricRepository) MetricUpdater {
	return &MetricService{repo: repo}
}

var _ MetricUpdater = (*MetricService)(nil)

func (s *MetricService) MetricUpdate(metricType string, metricName string, metricValue string) error {
	metric := s.repo.GetMetric(metricType, metricName)
	if metric == nil {
		metric = model.CreateMetric(metricName, metricType)
	}

	if err := metric.UpdateMetric(metricValue); err != nil {
		return err
	}
	s.repo.Save(metric)
	return nil
}
