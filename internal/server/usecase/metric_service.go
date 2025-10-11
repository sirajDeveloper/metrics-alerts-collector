package usecase

import (
	"fmt"
	"strconv"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

type MetricUpdater interface {
	MetricUpdate(metricType string, metricName string, metricValue string) error
}

type MetricGetter interface {
	GetMetricValue(metricType, metricName string) (string, error)
	GetAllMetricsForDisplay() []dto.MetricDTO
}

type MetricService struct {
	repo repository.MetricRepository
}

func NewMetricService(repo repository.MetricRepository) *MetricService {
	return &MetricService{repo: repo}
}

var _ MetricUpdater = (*MetricService)(nil)
var _ MetricGetter = (*MetricService)(nil)

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

func (s *MetricService) GetMetricValue(metricType, metricName string) (string, error) {
	metric := s.repo.GetMetric(metricType, metricName)
	if metric == nil {
		return "", fmt.Errorf("metric not found")
	}

	return formatMetricValue(metric, metricType)
}

func (s *MetricService) GetAllMetricsForDisplay() []dto.MetricDTO {
	metrics := s.repo.GetAll()

	displayMetrics := make([]dto.MetricDTO, 0, len(metrics))
	for _, m := range metrics {
		var valueStr string
		switch m.MType {
		case "gauge":
			if m.Value != nil {
				valueStr = strconv.FormatFloat(*m.Value, 'f', -1, 64)
			}
		case "counter":
			if m.Delta != nil {
				valueStr = strconv.FormatInt(*m.Delta, 10)
			}
		}

		displayMetrics = append(displayMetrics, dto.MetricDTO{
			ID:       m.ID,
			MType:    m.MType,
			ValueStr: valueStr,
		})
	}

	return displayMetrics
}

func formatMetricValue(metric *model.Metrics, metricType string) (string, error) {
	switch metricType {
	case "gauge":
		if metric.Value != nil {
			return strconv.FormatFloat(*metric.Value, 'f', -1, 64), nil
		}
	case "counter":
		if metric.Delta != nil {
			return strconv.FormatInt(*metric.Delta, 10), nil
		}
	}

	return "", fmt.Errorf("metric value is nil")
}
