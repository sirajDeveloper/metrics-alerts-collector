package bootstrap

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type UseCaseInitializer struct {
	config         Config
	metricRepo     repository.MetricRepository
	fileStorage    repository.MetricFileStorage
	healthChecker  usecase.DatabaseHealthChecker
	auditPublisher event.AuditEventPublisher
}

func NewUseCaseInitializer(cfg Config, metricRepo repository.MetricRepository, fileStorage repository.MetricFileStorage, healthChecker usecase.DatabaseHealthChecker, auditPublisher event.AuditEventPublisher) *UseCaseInitializer {
	return &UseCaseInitializer{
		config:         cfg,
		metricRepo:     metricRepo,
		fileStorage:    fileStorage,
		healthChecker:  healthChecker,
		auditPublisher: auditPublisher,
	}
}

type UseCaseResult struct {
	MetricUpdater usecase.MetricUpdater
	MetricGetter  usecase.MetricGetter
	HealthService usecase.HealthChecker
	Emitter       *usecase.MetricsEmitterService
}

func (u *UseCaseInitializer) Initialize() *UseCaseResult {
	emitter := usecase.NewMetricsEmitterService(u.fileStorage, u.metricRepo, *u.config.GetStoreInterval())
	metricService := usecase.NewMetricService(u.metricRepo, emitter, u.auditPublisher)
	healthService := usecase.NewHealthService(u.healthChecker)

	return &UseCaseResult{
		MetricUpdater: metricService,
		MetricGetter:  metricService,
		HealthService: healthService,
		Emitter:       emitter,
	}
}
