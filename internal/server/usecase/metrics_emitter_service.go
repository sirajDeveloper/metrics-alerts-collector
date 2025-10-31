package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type MetricsEmitterService struct {
	fStorage       repository.MetricFileStorage
	mRepo          repository.MetricRepository
	reportInterval int
}

var _ event.MetricsSender = (*MetricsEmitterService)(nil)

func NewMetricsEmitterService(fStorage repository.MetricFileStorage, mRepo repository.MetricRepository, reportInterval int) *MetricsEmitterService {
	return &MetricsEmitterService{
		fStorage:       fStorage,
		mRepo:          mRepo,
		reportInterval: reportInterval}
}

func (s *MetricsEmitterService) Send(e event.MetricsEvent) {
	logger.Log.Info("MetricsEmitterService.Send start")
	if s.reportInterval == 0 {
		s.mRepo.Save(e.Metrics)
	}
	logger.Log.Info("MetricsEmitterService.Send end")
}

func (s *MetricsEmitterService) EmitAll() {
	logger.Log.Info("MetricsEmitterService.EmitAll start")
	metrics := s.mRepo.GetAll()
	s.fStorage.SaveAll(metrics)
	logger.Log.Info("MetricsEmitterService.EmitAll end", zap.Int("count", len(metrics)))
}

func (s *MetricsEmitterService) RestoreAll() error {
	logger.Log.Info("MetricsEmitterService.RestoreAll start")
	metrics, err := s.fStorage.LoadAll()
	if err != nil {
		logger.Log.Error("MetricsEmitterService.RestoreAll load error", zap.Error(err))
		return err
	}
	for _, metric := range metrics {
		s.mRepo.Save(metric)
	}
	logger.Log.Info("MetricsEmitterService.RestoreAll end", zap.Int("count", len(metrics)))
	return nil
}
