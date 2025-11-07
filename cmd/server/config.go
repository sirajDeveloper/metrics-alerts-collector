package main

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

func parseConfig() (*Config, error) {
	var flagAddress string
	var flagStoreInterval int
	var flagFileStoragePath string
	var flagRestore bool
	var flagDatabaseDSN string

	flag.StringVar(&flagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval seconds (0 = sync)")
	flag.StringVar(&flagFileStoragePath, "f", "./metrics.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore saved values on start")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database connection string")
	flag.Parse()

	_ = godotenv.Load(".env")

	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Address == nil {
		cfg.Address = &flagAddress
	}
	if cfg.StoreInterval == nil {
		cfg.StoreInterval = &flagStoreInterval
	}
	if cfg.FileStoragePath == nil {
		cfg.FileStoragePath = &flagFileStoragePath
	}
	if cfg.Restore == nil {
		cfg.Restore = &flagRestore
	}
	if (cfg.DatabaseDSN == nil || *cfg.DatabaseDSN == "") && flagDatabaseDSN != "" {
		cfg.DatabaseDSN = &flagDatabaseDSN
	}
	if cfg.DatabaseDSN == nil || *cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("database dsn is required")
	}

	return cfg, nil
}

type Config struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     *string `env:"DATABASE_DSN"`
}
