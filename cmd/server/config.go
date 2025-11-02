package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

func parseConfig() (*Config, error) {
	var flagAddress string
	var flagStoreInterval int
	var flagFileStoragePath string
	var flagRestore bool

	flag.StringVar(&flagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval seconds (0 = sync)")
	flag.StringVar(&flagFileStoragePath, "f", "./metrics.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore saved values on start")
	flag.Parse()

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

	return cfg, nil
}

type Config struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
}
