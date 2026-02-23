package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
)

type serverFileConfig struct {
	Address         string `json:"address"`
	StoreInterval   int    `json:"store_interval"`
	FileStoragePath string `json:"file_storage_path"`
	Restore         bool   `json:"restore"`
	DatabaseDSN     string `json:"database_dsn"`
	MigrationsPath  string `json:"migrations_path"`
	CountRetrySave  int    `json:"count_retry_save"`
	SecretKey       string `json:"secret_key"`
	AuditFilePath   string `json:"audit_file_path"`
	AuditServiceURL string `json:"audit_service_url"`
	CryptoKey       string `json:"crypto_key"`
	EnableHTTPS     bool   `json:"enable_https"`
	TLSCertFile     string `json:"tls_cert_file"`
	TLSKeyFile      string `json:"tls_key_file"`
}

// parseConfig loads server configuration from file, environment and flags.
// Priority (lowest to highest): file < env < flags.
func parseConfig() (*Config, error) {
	_ = godotenv.Load(".env")

	cfg, err := loadConfigFromFile(getConfigPath())
	if err != nil {
		return nil, err
	}
	if err = applyConfigFromEnv(cfg); err != nil {
		return nil, err
	}
	applyConfigFromFlags(cfg)
	return cfg, nil
}

// getConfigPath returns config file path from flag -c/-config or env CONFIG.
func getConfigPath() string {
	p := os.Getenv("CONFIG")
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringVar(&p, "c", p, "")
	fs.StringVar(&p, "config", p, "")
	_ = fs.Parse(os.Args[1:])
	return p
}

// loadConfigFromFile reads JSON from path and returns Config with file values set.
// Returns empty Config if path is empty. Returns error on read or invalid JSON.
func loadConfigFromFile(path string) (*Config, error) {
	if path == "" {
		return &Config{}, nil
	}
	fc, err := readServerFileConfig(path)
	if err != nil {
		return nil, err
	}
	return configFromFileStruct(fc), nil
}

// readServerFileConfig reads and unmarshals JSON config file. Returns zero value if path empty.
func readServerFileConfig(path string) (serverFileConfig, error) {
	var fc serverFileConfig
	if path == "" {
		return fc, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fc, err
	}
	if err := json.Unmarshal(data, &fc); err != nil {
		return fc, err
	}
	return fc, nil
}

// configFromFileStruct builds Config from parsed file config struct.
func configFromFileStruct(fc serverFileConfig) *Config {
	cfg := &Config{}
	if fc.Address != "" {
		cfg.Address = &fc.Address
	}
	cfg.StoreInterval = &fc.StoreInterval
	if fc.FileStoragePath != "" {
		cfg.FileStoragePath = &fc.FileStoragePath
	}
	cfg.Restore = &fc.Restore
	if fc.DatabaseDSN != "" {
		cfg.DatabaseDSN = &fc.DatabaseDSN
		cfg.MigrationsPath = &fc.MigrationsPath
	}
	if fc.CountRetrySave != 0 {
		cfg.CountRetrySave = &fc.CountRetrySave
	}
	if fc.SecretKey != "" {
		cfg.SecretKey = &fc.SecretKey
	}
	if fc.AuditFilePath != "" {
		cfg.AuditFilePath = &fc.AuditFilePath
	}
	if fc.AuditServiceURL != "" {
		cfg.AuditServiceURL = &fc.AuditServiceURL
	}
	if fc.CryptoKey != "" {
		cfg.CryptoKey = &fc.CryptoKey
	}
	cfg.EnableHTTPS = &fc.EnableHTTPS
	if fc.TLSCertFile != "" {
		cfg.TLSCertFile = &fc.TLSCertFile
	}
	if fc.TLSKeyFile != "" {
		cfg.TLSKeyFile = &fc.TLSKeyFile
	}
	return cfg
}

// applyConfigFromEnv overwrites cfg fields with environment variables (ADDRESS, KEY, etc.).
// Returns error if env parsing fails.
func applyConfigFromEnv(cfg *Config) error {
	return env.Parse(cfg)
}

// applyConfigFromFlags defines flags, parses os.Args, and sets cfg fields that are still nil
// from flag values (flags have highest priority).
func applyConfigFromFlags(cfg *Config) {
	var flagAddress string
	var flagStoreInterval int
	var flagFileStoragePath string
	var flagRestore bool
	var flagDatabaseDSN string
	var flagMigrationsPath string
	var flagCountRetrySave int
	var flagSecretKey string
	var flagAuditFilePath string
	var flagAuditServiceURL string
	var flagCryptoKey string
	var flagEnableHTTPS bool
	var flagTLSCertFile string
	var flagTLSKeyFile string

	var configPathDummy string
	flag.StringVar(&configPathDummy, "c", "", "config file path (JSON)")
	flag.StringVar(&configPathDummy, "config", "", "config file path (JSON)")
	flag.StringVar(&flagAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagStoreInterval, "i", 300, "store interval seconds (0 = sync)")
	flag.StringVar(&flagFileStoragePath, "f", "./metrics.json", "file storage path")
	flag.BoolVar(&flagRestore, "r", true, "restore saved values on start")
	flag.StringVar(&flagDatabaseDSN, "d", "", "database connection string")
	flag.StringVar(&flagMigrationsPath, "m", "./migrations", "migrations directory")
	flag.IntVar(&flagCountRetrySave, "retry", 3, "count of retry attempts for database save")
	flag.StringVar(&flagSecretKey, "k", "", "secret key for signature")
	flag.StringVar(&flagAuditFilePath, "audit-file", "./audit/audit.log", "audit file path")
	flag.StringVar(&flagAuditServiceURL, "audit-service", "", "external audit service URL")
	flag.StringVar(&flagCryptoKey, "crypto-key", "", "path to private key file for decryption")
	flag.BoolVar(&flagEnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&flagTLSCertFile, "tls-cert", "./server.crt", "path to TLS certificate file")
	flag.StringVar(&flagTLSKeyFile, "tls-key", "./server.key", "path to TLS private key file")
	flag.Parse()

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
	cfg.AuditFilePath = optionalString(cfg.AuditFilePath, flagAuditFilePath)
	cfg.AuditServiceURL = optionalString(cfg.AuditServiceURL, flagAuditServiceURL)
	cfg.CryptoKey = optionalString(cfg.CryptoKey, flagCryptoKey)
	if cfg.EnableHTTPS == nil {
		cfg.EnableHTTPS = &flagEnableHTTPS
	}
	cfg.TLSCertFile = optionalString(cfg.TLSCertFile, flagTLSCertFile)
	cfg.TLSKeyFile = optionalString(cfg.TLSKeyFile, flagTLSKeyFile)
}

// Config holds server configuration. Fields are pointers so env/flag can leave them unset.
// See parseConfig for load order: file < env < flags.
type Config struct {
	Address         *string `env:"ADDRESS"`
	StoreInterval   *int    `env:"STORE_INTERVAL"`
	FileStoragePath *string `env:"FILE_STORAGE_PATH"`
	Restore         *bool   `env:"RESTORE"`
	DatabaseDSN     *string `env:"DATABASE_DSN"`
	MigrationsPath  *string `env:"MIGRATIONS_PATH"`
	CountRetrySave  *int    `env:"COUNT_RETRY_SAVE"`
	SecretKey       *string `env:"KEY"`
	AuditFilePath   *string `env:"AUDIT_FILE_PATH"`
	AuditServiceURL *string `env:"AUDIT_SERVICE_URL"`
	CryptoKey       *string `env:"CRYPTO_KEY"`
	EnableHTTPS     *bool   `env:"ENABLE_HTTPS"`
	TLSCertFile     *string `env:"TLS_CERT_FILE"`
	TLSKeyFile      *string `env:"TLS_KEY_FILE"`
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

func (c *Config) GetAuditFilePath() *string {
	return c.AuditFilePath
}

func (c *Config) GetAuditServiceURL() *string {
	return c.AuditServiceURL
}

func (c *Config) GetCryptoKey() *string {
	return c.CryptoKey
}

func (c *Config) GetEnableHTTPS() *bool {
	return c.EnableHTTPS
}

func (c *Config) GetTLSCertFile() *string {
	return c.TLSCertFile
}

func (c *Config) GetTLSKeyFile() *string {
	return c.TLSKeyFile
}

// optionalString returns current if non-nil and non-empty, else pointer to fallback or nil if fallback empty.
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
