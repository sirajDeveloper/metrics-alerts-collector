package bootstrap

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/scheduler"
)

type App struct {
	config      Config
	server      *http.Server
	scheduler   *scheduler.MetricEmitterScheduler
	schedCtx    context.Context
	schedCancel context.CancelFunc
	db          *sqlx.DB
}

func NewApp(cfg Config) *App {
	return &App{
		config: cfg,
	}
}

func (a *App) Initialize() error {
	infrastructureInitializer := NewInfrastructureInitializer(a.config)
	infraResult, err := infrastructureInitializer.Initialize()
	if err != nil {
		return err
	}

	a.db = infraResult.DB

	useCaseInitializer := NewUseCaseInitializer(a.config, infraResult.MetricRepository, infraResult.FileStorage, infraResult.HealthChecker)
	useCaseResult := useCaseInitializer.Initialize()

	handlerInitializer := NewHandlerInitializer(a.config, useCaseResult.MetricUpdater, useCaseResult.MetricGetter, useCaseResult.HealthService, useCaseResult.Emitter)
	handlerResult := handlerInitializer.Initialize()

	a.server = handlerResult.Server
	a.scheduler = handlerResult.Scheduler
	a.schedCtx = handlerResult.SchedCtx
	a.schedCancel = handlerResult.SchedCancel

	return nil
}

func (a *App) Run() error {
	go func() {
		logger.Log.Info("Server starting on http://" + *a.config.GetAddress())
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Server failed to start", zap.String("error", err.Error()))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	return a.Shutdown()
}

func (a *App) Shutdown() error {
	logger.Log.Info("Shutting down gracefully...")
	a.scheduler.Shutdown()
	a.schedCancel()

	if a.db != nil {
		defer a.db.Close()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server shutdown failed", zap.String("error", err.Error()))
		return err
	}

	logger.Log.Info("Server stopped")
	return nil
}
