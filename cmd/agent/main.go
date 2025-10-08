package main

import (
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/infrastructure"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
)

func main() {
	sender := infrastructure.NewHTTPSender("http://localhost:8080")
	collector := usecase.NewCollector(sender)

	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second

	go func() {
		for {
			collector.Collect()
			time.Sleep(pollInterval)
		}
	}()

	go func() {
		for {
			collector.Report()
			time.Sleep(reportInterval)
		}
	}()

	select {}
}
