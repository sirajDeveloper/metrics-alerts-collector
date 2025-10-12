package main

import (
	"fmt"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/infrastructure"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
)

func main() {
	ParseFlags()
	serverURL := "http://localhost" + Port
	sender := infrastructure.NewHTTPSender(serverURL)
	fmt.Printf("HTTPSender init with serverURL: %v", serverURL)
	collector := usecase.NewCollector(sender)

	go func() {
		for {
			collector.Collect()
			time.Sleep(time.Duration(PollInterval) * time.Second)
		}
	}()

	go func() {
		for {
			collector.Report()
			time.Sleep(time.Duration(ReportInterval) * time.Second)
		}
	}()

	select {}
}
