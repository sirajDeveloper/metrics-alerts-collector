package main

import (
	"flag"
)

var address string

func parseFlags() {
	flag.StringVar(&address, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}
