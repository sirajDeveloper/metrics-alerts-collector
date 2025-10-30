package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
)

var address string
var storeInterval int
var fileStoragePath string
var restore bool

func parseConfig() {
	flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&storeInterval, "i", 300, "store interval seconds (0 = sync)")
	flag.StringVar(&fileStoragePath, "f", "./metrics.json", "file storage path")
	flag.BoolVar(&restore, "r", false, "restore saved values on start")
	flag.Parse()
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		logger.Log.Fatal("Failed to parse config", zap.String("error", err.Error()))
	}
	if cfg.Address != "" {
		address = cfg.Address
	}
	if cfg.StoreInterval != nil {
		storeInterval = *cfg.StoreInterval
	}
	if cfg.FileStoragePath != "" {
		fileStoragePath = cfg.FileStoragePath
	}
	if cfg.Restore != nil {
		restore = *cfg.Restore
	}
}

type Config struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   *int   `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
}
