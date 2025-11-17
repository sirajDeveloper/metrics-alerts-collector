package bootstrap

import (
	"context"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/scheduler"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type HandlerInitializer struct {
	config        Config
	metricUpdater usecase.MetricUpdater
	metricGetter  usecase.MetricGetter
	healthService usecase.HealthChecker
	emitter       *usecase.MetricsEmitterService
}

func NewHandlerInitializer(cfg Config, metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter, healthService usecase.HealthChecker, emitter *usecase.MetricsEmitterService) *HandlerInitializer {
	return &HandlerInitializer{
		config:        cfg,
		metricUpdater: metricUpdater,
		metricGetter:  metricGetter,
		healthService: healthService,
		emitter:       emitter,
	}
}

type HandlerResult struct {
	Server      *http.Server
	Scheduler   *scheduler.MetricEmitterScheduler
	SchedCtx    context.Context
	SchedCancel context.CancelFunc
}

func (h *HandlerInitializer) Initialize() *HandlerResult {
	schedulerInstance := scheduler.NewMetricEmitterScheduler(h.emitter, *h.config.GetStoreInterval(), *h.config.GetRestore())
	schedCtx, schedCancel := context.WithCancel(context.Background())
	schedulerInstance.Start(schedCtx)

	chiRouter := router.NewChiRouter(h.metricUpdater, h.metricGetter, h.healthService)
	server := &http.Server{
		Addr:    *h.config.GetAddress(),
		Handler: chiRouter.Handler(),
	}

	return &HandlerResult{
		Server:      server,
		Scheduler:   schedulerInstance,
		SchedCtx:    schedCtx,
		SchedCancel: schedCancel,
	}
}
