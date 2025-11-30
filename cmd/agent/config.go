package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

var (
	address        string
	reportInterval int
	pollInterval   int
	countRetrySave int
	secretKey      string
	rateLimit      int
)

func ParseConfig() {
	flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.IntVar(&countRetrySave, "retry", 3, "count of retry attempts for database save")
	flag.StringVar(&secretKey, "k", "", "secret key for signature")
	flag.IntVar(&rateLimit, "l", 1, "rate limit for concurrent requests")
	flag.Parse()
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	if cfg.Address != "" {
		address = cfg.Address
	}
	if cfg.PollInterval != 0 {
		pollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval != 0 {
		reportInterval = cfg.ReportInterval
	}
	if cfg.CountRetrySave != 0 {
		countRetrySave = cfg.CountRetrySave
	}
	if cfg.SecretKey != "" {
		secretKey = cfg.SecretKey
	}
	if cfg.RateLimit != 0 {
		rateLimit = cfg.RateLimit
	}
}

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	CountRetrySave int    `env:"COUNT_RETRY_SAVE"`
	SecretKey      string `env:"KEY"`
	RateLimit      int    `env:"RATE_LIMIT"`
}
