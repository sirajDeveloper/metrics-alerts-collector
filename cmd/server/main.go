package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"

	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/database"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/file"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/scheduler"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

func main() {
	logger.InitLogger(false)
	cfg, err := parseConfig()
	if err != nil {
		logger.Log.Fatal("Failed to parse config", zap.String("error", err.Error()))
	}

	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Log.Info("Failed to sync logs")
		}
	}()

	fileStorage := file.NewJSONFileStorage(*cfg.FileStoragePath)

	var metricRepo repository.MetricRepository = cache.NewMemStorage()

	var healthChecker usecase.DatabaseHealthChecker
	if cfg.DatabaseDSN != nil {
		if cfg.MigrationsPath != nil {
			if migrationRunner, migErr := database.NewMigrationRunner(*cfg.MigrationsPath, *cfg.DatabaseDSN); migErr != nil {
				logger.Log.Error("Failed to initialize migrations", zap.Error(migErr))
			} else if migErr = migrationRunner.Up(context.Background()); migErr != nil {
				logger.Log.Error("Failed to apply migrations", zap.Error(migErr))
			}
		}

		dbCtx, cancelDB := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelDB()

		poolConfig, err := pgxpool.ParseConfig(*cfg.DatabaseDSN)
		if err != nil {
			logger.Log.Error("Failed to parse database config", zap.Error(err))
		} else {
			pool, err := pgxpool.NewWithConfig(dbCtx, poolConfig)
			if err != nil {
				logger.Log.Error("Failed to initialize database", zap.Error(err))
			} else {
				defer pool.Close()

				mPostgresRepo := database.NewMetricsPostgresRepository(pool)
				metricRepo = mPostgresRepo
				healthChecker = database.NewDBhealthCheckImpl(pool)
			}
		}
	}

	emitter := usecase.NewMetricsEmitterService(fileStorage, metricRepo, *cfg.StoreInterval)

	metricService := usecase.NewMetricService(metricRepo, emitter)
	healthService := usecase.NewHealthService(healthChecker)

	emitStarter := scheduler.NewMetricEmitterScheduler(emitter, *cfg.StoreInterval, *cfg.Restore)

	schedCtx, schedCancel := context.WithCancel(context.Background())
	defer schedCancel()
	emitStarter.Start(schedCtx)

	chiRouter := router.NewChiRouter(metricService, metricService, healthService)

	server := &http.Server{
		Addr:    *cfg.Address,
		Handler: chiRouter.Handler(),
	}

	go func() {
		logger.Log.Info("Server starting on http://" + *cfg.Address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Server failed to start", zap.String("error", err.Error()))
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	logger.Log.Info("Shutting down gracefully...")
	emitStarter.Shutdown()

	ctxt, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()
	if err := server.Shutdown(ctxt); err != nil {
		logger.Log.Fatal("Server shutdown failed", zap.String("error", err.Error()))
	}
	logger.Log.Info("Server stopped")
}
