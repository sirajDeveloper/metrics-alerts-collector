package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

func parseConfig() (*Config, error) {
	var flagAddress string
	var flagStoreInterval int
	var flagFileStoragePath string
	var flagRestore bool
	var flagDatabaseDSN string
	var flagMigrationsPath string
	var flagCountRetrySave int
	var flagSecretKey string

	flag.StringVar(&flagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval seconds (0 = sync)")
	flag.StringVar(&flagFileStoragePath, "f", "./metrics.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore saved values on start")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&flagMigrationsPath, "m", "./migrations", "migrations directory")
	flag.IntVar(&flagCountRetrySave, "retry", 3, "count of retry attempts for database save")
	flag.StringVar(&flagSecretKey, "k", "", "secret key for signature")
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
	cfg.DatabaseDSN = optionalString(cfg.DatabaseDSN, flagDatabaseDSN)
	if cfg.DatabaseDSN != nil {
		cfg.MigrationsPath = optionalString(cfg.MigrationsPath, flagMigrationsPath)
	} else {
		cfg.MigrationsPath = nil
	}
	if cfg.CountRetrySave == nil {
		cfg.CountRetrySave = &flagCountRetrySave
	}
	if cfg.SecretKey == nil {
		cfg.SecretKey = &flagSecretKey
	}

	return cfg, nil
}

type Config struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     *string `env:"DATABASE_DSN"`
	MigrationsPath  *string `env:"MIGRATIONS_PATH"`
	CountRetrySave  *int    `env:"COUNT_RETRY_SAVE"`
	SecretKey       *string `env:"KEY"`
}

func (c *Config) GetAddress() *string {
	return c.Address
}

func (c *Config) GetStoreInterval() *int {
	return c.StoreInterval
}

func (c *Config) GetFileStoragePath() *string {
	return c.FileStoragePath
}

func (c *Config) GetRestore() *bool {
	return c.Restore
}

func (c *Config) GetDatabaseDSN() *string {
	return c.DatabaseDSN
}

func (c *Config) GetMigrationsPath() *string {
	return c.MigrationsPath
}

func (c *Config) GetCountRetrySave() *int {
	return c.CountRetrySave
}
func (c *Config) GetSecretKey() *string {
	return c.SecretKey
}

func optionalString(current *string, fallback string) *string {
	if current != nil && *current != "" {
		return current
	}
	if fallback == "" {
		return nil
	}
	value := fallback
	return &value
}
