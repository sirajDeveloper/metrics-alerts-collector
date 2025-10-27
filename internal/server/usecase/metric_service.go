package usecase

import (
	"fmt"
	"strconv"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

type MetricUpdater interface {
	MetricUpdate(req *dto.MetricUpdateRequest) error
}

type MetricGetter interface {
	GetMetricValue(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error)
	GetAllMetricsForDisplay() []dto.DisplayMetricDTO
}

var _ MetricUpdater = (*MetricService)(nil)
var _ MetricGetter = (*MetricService)(nil)

type MetricService struct {
	repo repository.MetricRepository
}

func NewMetricService(repo repository.MetricRepository) *MetricService {
	return &MetricService{repo: repo}
}

func (s *MetricService) MetricUpdate(req *dto.MetricUpdateRequest) error {
	metric := s.repo.GetMetric(req.Type, req.Name)
	if metric == nil {
		metric = model.CreateMetric(req.Name, req.Type)
	}

	if err := metric.UpdateMetric(req.Value); err != nil {
		return err
	}
	s.repo.Save(metric)
	return nil
}

func (s *MetricService) GetMetricValue(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
	metric := s.repo.GetMetric(req.Type, req.Name)
	resp := dto.MetricValueResponse{
		ID:    "",
		MType: "",
		Delta: nil,
		Value: nil,
	}
	if metric == nil {
		return &resp, fmt.Errorf("metric not found")
	}

	resp.ID = metric.ID
	resp.MType = metric.MType

	switch metric.MType {
	case "gauge":
		if metric.Value != nil {
			resp.Value = metric.Value
		}
	case "counter":
		if metric.Delta != nil {
			resp.Delta = metric.Delta
		}
	}

	return &resp, nil
}

func (s *MetricService) GetAllMetricsForDisplay() []dto.DisplayMetricDTO {
	metrics := s.repo.GetAll()

	displayMetrics := make([]dto.DisplayMetricDTO, 0, len(metrics))
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

		displayMetrics = append(displayMetrics, dto.DisplayMetricDTO{
			ID:       m.ID,
			MType:    m.MType,
			ValueStr: valueStr,
		})
	}

	return displayMetrics
}
