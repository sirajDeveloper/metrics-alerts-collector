// @title Metrics Alerts Collector API
// @version 1.0
// @description API сервера для сбора метрик и алертинга
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http https
package main

import (
	"fmt"
	"os"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/bootstrap"
	"go.uber.org/zap"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()
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

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Fprintf(os.Stdout, "Build version: %s\n", version)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", date)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", commit)
}
