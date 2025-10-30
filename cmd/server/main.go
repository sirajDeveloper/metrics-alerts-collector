package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/file"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/scheduler"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

func main() {
	parseConfig()

	logger.InitLogger(false)
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Log.Info("Failed to sync logs")
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fileStorage := file.NewJSONFileStorage(fileStoragePath)
	emitter := usecase.NewMetricsEmitterService(fileStorage, storeInterval)

	emitStarter := scheduler.NewMetricEmitterScheduler(emitter, storeInterval, restore)

	emitStarter.Start(ctx)

	metricRepo := cache.NewMemStorage()
	metricService := usecase.NewMetricService(metricRepo, emitter)
	chiRouter := router.NewChiRouter(metricService, metricService)

	server := &http.Server{
		Addr:    address,
		Handler: chiRouter.Handler(),
	}

	go func() {
		logger.Log.Info("Server starting on http://" + address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Server failed to start", zap.String("error", err.Error()))
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Log.Info("Shutting down gracefully...")

	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Shutdown(ctxt); err != nil {
		logger.Log.Fatal("Server shutdown failed", zap.String("error", err.Error()))
	}
	time.Sleep(100 * time.Millisecond)
	logger.Log.Info("Server stopped")
}
