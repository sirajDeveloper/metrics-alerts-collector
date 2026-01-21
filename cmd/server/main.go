package main

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/bootstrap"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger(false)

	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Log.Info("Failed to sync logs")
		}
	}()

	cfg, err := parseConfig()
	if err != nil {
		logger.Log.Fatal("Failed to parse config", zap.String("error", err.Error()))
	}

	app := bootstrap.NewApp(cfg)
	if err := app.Initialize(); err != nil {
		logger.Log.Fatal("Failed to initialize app", zap.String("error", err.Error()))
	}

	if err := app.Run(); err != nil {
		logger.Log.Fatal("App failed", zap.String("error", err.Error()))
	}
}
