package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"log"
)

var (
	address        string
	reportInterval int
	pollInterval   int
)

func ParseConfig() {
	flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
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
}

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}
