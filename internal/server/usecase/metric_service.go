package usecase

import (
	"fmt"
	"strconv"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
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
	repo   repository.MetricRepository
	sender event.MetricsSender
}

func NewMetricService(repo repository.MetricRepository, sender event.MetricsSender) *MetricService {
	return &MetricService{repo: repo, sender: sender}
}

func (s *MetricService) MetricUpdate(req *dto.MetricUpdateRequest) error {
	metric := s.repo.GetMetric(req.MType, req.ID)
	if metric == nil {
		metric = model.CreateMetric(req.ID, req.MType)
	}

	var value any
	switch req.MType {
	case "gauge":
		if req.Value != nil {
			value = *req.Value
		} else {
			return fmt.Errorf("gauge value is required")
		}
	case "counter":
		if req.Delta != nil {
			value = *req.Delta
		} else {
			return fmt.Errorf("counter delta is required")
		}
	default:
		return fmt.Errorf("unknown metric type: %s", req.MType)
	}

	if err := metric.UpdateMetric(value); err != nil {
		return err
	}
	s.repo.Save(metric)
	s.putEvent(metric)
	return nil
}

func (s *MetricService) putEvent(metric *model.Metrics) {
	if s.sender == nil || metric == nil {
		return
	}
	s.sender.Send(event.MetricsEvent{Metrics: metric})
}

func (s *MetricService) GetMetricValue(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
	metric := s.repo.GetMetric(req.MType, req.ID)
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
