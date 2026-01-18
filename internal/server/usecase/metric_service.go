package usecase

import (
	"fmt"
	"strconv"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

// MetricUpdater определяет интерфейс для обновления метрик.
// Реализуется MetricService.
type MetricUpdater interface {
	// MetricUpdate обновляет или создает метрику.
	MetricUpdate(req *dto.MetricUpdateRequest) error
}

// MetricGetter определяет интерфейс для получения метрик.
// Реализуется MetricService.
type MetricGetter interface {
	// GetMetricValue возвращает значение метрики по имени и типу.
	GetMetricValue(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error)
	// GetAllMetricsForDisplay возвращает все метрики для отображения.
	GetAllMetricsForDisplay() []dto.DisplayMetricDTO
}

var _ MetricUpdater = (*MetricService)(nil)
var _ MetricGetter = (*MetricService)(nil)

// MetricService реализует бизнес-логику работы с метриками.
// Обеспечивает создание, обновление и получение метрик.
// Реализует интерфейсы MetricUpdater и MetricGetter.
type MetricService struct {
	repo   repository.MetricRepository
	sender event.MetricsSender
}

// NewMetricService создает новый экземпляр MetricService.
//
// Параметры:
//   - repo: репозиторий для хранения метрик
//   - sender: отправитель событий о метриках (может быть nil)
//
// Возвращает новый экземпляр MetricService.
func NewMetricService(repo repository.MetricRepository, sender event.MetricsSender) *MetricService {
	return &MetricService{repo: repo, sender: sender}
}

// MetricUpdate обновляет или создает метрику.
//
// Для gauge: устанавливает абсолютное значение.
// Для counter: добавляет delta к текущему значению.
//
// Параметры:
//   - req: запрос на обновление метрики
//
// Возвращает ошибку, если тип метрики неизвестен или данные некорректны.
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

// GetMetricValue возвращает значение метрики по имени и типу.
//
// Параметры:
//   - req: запрос на получение метрики
//
// Возвращает ответ с данными метрики или ошибку, если метрика не найдена.
func (s *MetricService) GetMetricValue(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
	metric := s.repo.GetMetric(req.MType, req.ID)
	if metric == nil {
		return nil, fmt.Errorf("metric not found")
	}

	resp := &dto.MetricValueResponse{
		ID:    metric.ID,
		MType: metric.MType,
		Delta: nil,
		Value: nil,
	}

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

	return resp, nil
}

// GetAllMetricsForDisplay возвращает все метрики в формате для отображения.
//
// Возвращает слайс DTO с метриками, где значения представлены в виде строк.
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
