package main

import (
	"flag"
)

var (
	Port           string
	ReportInterval int
	PollInterval   int
)

func ParseFlags() {
	flag.StringVar(&Port, "port", ":8080", "address and port to run server")
	flag.IntVar(&PollInterval, "pollInterval", 30, "poll interval in seconds")
	flag.IntVar(&ReportInterval, "reportInterval", 45, "report interval in seconds")
	flag.Parse()
}
