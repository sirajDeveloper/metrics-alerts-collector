package application

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain"
)

type MetricUpdater interface {
	MetricUpdate(metricType string, metricName string, metricValue string) error
}

type MetricService struct {
	repo domain.MetricRepository
}

func NewMetricService(repo domain.MetricRepository) MetricUpdater {
	return &MetricService{repo: repo}
}

var _ MetricUpdater = (*MetricService)(nil)

func (s *MetricService) MetricUpdate(metricType string, metricName string, metricValue string) error {
	metric := s.repo.GetMetric(metricType, metricName)
	metric.MType = metricType
	if err := metric.UpdateMetric(metricValue); err != nil {
		return err
	}
	s.repo.Save(metric)
	return nil
}
