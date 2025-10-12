package main

import (
	"flag"
)

var port string

func parseFlags() {
	flag.StringVar(&port, "a", ":8080", "address and port to run server")
	flag.Parse()
}
