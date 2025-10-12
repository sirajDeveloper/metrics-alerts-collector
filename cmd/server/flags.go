package main

import (
	"flag"
)

var port string

func parseFlags() {
	flag.StringVar(&port, "port", ":8080", "address and port to run server")
	flag.Parse()
}
