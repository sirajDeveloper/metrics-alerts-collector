package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
)

var address string

func parseConfig() {
	flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		logger.Log.Fatal("Failed to parse config", zap.String("error", err.Error()))
	}
	if cfg.Address != "" {
		address = cfg.Address
	}
}

type Config struct {
	Address string `env:"ADDRESS"`
}
