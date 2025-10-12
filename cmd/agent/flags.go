package main

import (
	"flag"
)

var (
	address        string
	reportInterval int
	pollInterval   int
)

func ParseFlags() {
	flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&pollInterval, "p", 2, "poll interval in seconds")
	flag.IntVar(&reportInterval, "r", 10, "report interval in seconds")
	flag.Parse()
}
