package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

var (
	address        string
	grpcAddress    string
	reportInterval int
	pollInterval   int
	countRetrySave int
	secretKey      string
	rateLimit      int
	cryptoKeyPath  string
	agentIP        string
)

type fileConfig struct {
	Address        string `json:"address"`
	GrpcAddress    string `json:"grpc_address"`
	ReportInterval int    `json:"report_interval"`
	PollInterval   int    `json:"poll_interval"`
	CountRetrySave int    `json:"count_retry_save"`
	SecretKey      string `json:"secret_key"`
	RateLimit      int    `json:"rate_limit"`
	CryptoKey      string `json:"crypto_key"`
	AgentIP        string `json:"agent_ip"`
}

// ParseConfig loads agent configuration from file, environment and flags.
// Priority (lowest to highest): file < env < flags.
func ParseConfig() {
	path := getConfigPath()
	fc := loadConfigFromFile(path)
	applyConfigFromFile(fc)
	applyConfigFromEnv()
	applyConfigFromFlags(path)
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

// loadConfigFromFile reads and parses JSON config from path.
// Returns zero value if path is empty. Exits on read or parse error.
func loadConfigFromFile(path string) fileConfig {
	var fc fileConfig
	if path == "" {
		return fc
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read config file: %v", err)
	}
	if err := json.Unmarshal(data, &fc); err != nil {
		log.Fatalf("parse config file: %v", err)
	}
	return fc
}

// applyConfigFromFile sets agent globals from file config, using defaults for empty/zero fields.
func applyConfigFromFile(fc fileConfig) {
	address = "localhost:8080"
	if fc.Address != "" {
		address = fc.Address
	}
	grpcAddress = "localhost:3200"
	if fc.GrpcAddress != "" {
		grpcAddress = fc.GrpcAddress
	}
	pollInterval = 2
	if fc.PollInterval != 0 {
		pollInterval = fc.PollInterval
	}
	reportInterval = 10
	if fc.ReportInterval != 0 {
		reportInterval = fc.ReportInterval
	}
	countRetrySave = 3
	if fc.CountRetrySave != 0 {
		countRetrySave = fc.CountRetrySave
	}
	secretKey = ""
	if fc.SecretKey != "" {
		secretKey = fc.SecretKey
	}
	rateLimit = 1
	if fc.RateLimit != 0 {
		rateLimit = fc.RateLimit
	}
	cryptoKeyPath = ""
	if fc.CryptoKey != "" {
		cryptoKeyPath = fc.CryptoKey
	}
	agentIP = "127.0.0.1"
	if fc.AgentIP != "" {
		agentIP = fc.AgentIP
	}
}

// applyConfigFromEnv overwrites agent globals with environment variables (ADDRESS, KEY, etc.).
func applyConfigFromEnv() {
	var envCfg Config
	if err := env.Parse(&envCfg); err != nil {
		log.Fatal(err)
	}
	if envCfg.Address != "" {
		address = envCfg.Address
	}
	if envCfg.GrpcAddress != "" {
		grpcAddress = envCfg.GrpcAddress
	}
	if envCfg.PollInterval != 0 {
		pollInterval = envCfg.PollInterval
	}
	if envCfg.ReportInterval != 0 {
		reportInterval = envCfg.ReportInterval
	}
	if envCfg.CountRetrySave != 0 {
		countRetrySave = envCfg.CountRetrySave
	}
	if envCfg.SecretKey != "" {
		secretKey = envCfg.SecretKey
	}
	if envCfg.RateLimit != 0 {
		rateLimit = envCfg.RateLimit
	}
	if envCfg.CryptoKey != "" {
		cryptoKeyPath = envCfg.CryptoKey
	}
	if envCfg.AgentIP != "" {
		agentIP = envCfg.AgentIP
	}
}

// applyConfigFromFlags defines flags with current globals as defaults, parses os.Args,
// and overwrites globals with flag values (highest priority).
// configPath is the path already resolved from -c/CONFIG for default flag value.
func applyConfigFromFlags(configPath string) {
	var configPathDummy string
	flag.StringVar(&configPathDummy, "c", configPath, "config file path (JSON)")
	flag.StringVar(&configPathDummy, "config", configPath, "config file path (JSON)")
	flag.StringVar(&address, "a", address, "address and port to run server")
	flag.StringVar(&grpcAddress, "grpc-address", grpcAddress, "gRPC server address")
	flag.IntVar(&pollInterval, "p", pollInterval, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", reportInterval, "report interval in seconds")
	flag.IntVar(&countRetrySave, "retry", countRetrySave, "count of retry attempts for database save")
	flag.StringVar(&secretKey, "k", secretKey, "secret key for signature")
	flag.IntVar(&rateLimit, "l", rateLimit, "rate limit for concurrent requests")
	flag.StringVar(&cryptoKeyPath, "crypto-key", cryptoKeyPath, "path to public key file for encryption")
	flag.StringVar(&agentIP, "agent-ip", agentIP, "agent IP address for gRPC metadata")
	flag.Parse()
}

// Config holds agent configuration parsed from environment (env tags).
type Config struct {
	Address        string `env:"ADDRESS"`
	GrpcAddress    string `env:"GRPC_ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	CountRetrySave int    `env:"COUNT_RETRY_SAVE"`
	SecretKey      string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
	CryptoKey      string `env:"CRYPTO_KEY"`
	AgentIP        string `env:"AGENT_IP"`
}
