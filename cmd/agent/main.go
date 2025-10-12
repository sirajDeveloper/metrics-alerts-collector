package main

import (
	"fmt"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/infrastructure"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
)

func main() {
	ParseFlags()
	serverURL := "http://" + address
	sender := infrastructure.NewHTTPSender(serverURL)
	fmt.Printf("HTTPSender init with serverURL: %v\n", serverURL)
	collector := usecase.NewCollector(sender)

	go func() {
		for {
			collector.Collect()
			time.Sleep(time.Duration(pollInterval) * time.Second)
		}
	}()

	go func() {
		for {
			collector.Report()
			time.Sleep(time.Duration(reportInterval) * time.Second)
		}
	}()

	select {}
}
