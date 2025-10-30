package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type MetricsEmitterService struct {
	repo           repository.MetricFileStorage
	getter         MetricGetter
	reportInterval int
}

var _ event.MetricsSender = (*MetricsEmitterService)(nil)

func NewMetricsEmitterService(repo repository.MetricFileStorage, getter MetricGetter, reportInterval int) *MetricsEmitterService {
	return &MetricsEmitterService{
		repo:           repo,
		getter:         getter,
		reportInterval: reportInterval}
}

func (s *MetricsEmitterService) Send(e event.MetricsEvent) {
	logger.Log.Info("MetricsEmitterService.Send start")
	if s.reportInterval == 0 {
		s.repo.Save(e.Metrics)
	}
	logger.Log.Info("MetricsEmitterService.Send end")
}

func (s *MetricsEmitterService) EmitAll() {
	logger.Log.Info("MetricsEmitterService.EmitAll start")
	metrics := s.getter.GetAllMetrics()
	s.repo.SaveAll(metrics)
	logger.Log.Info("MetricsEmitterService.EmitAll end", zap.Int("count", len(metrics)))
}
