package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/infrastructure"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
)

func main() {
	ParseConfig()
	logger.InitLogger(false)

	serverURL := "http://" + address
	sender := infrastructure.NewHTTPSender(serverURL, secretKey, countRetrySave)
	fmt.Printf("HTTPSender init with serverURL: %v\n", serverURL)
	fmt.Printf("Rate limit: %d concurrent requests\n", rateLimit)
	reporter := usecase.NewMetricWorkerPoolReporter(sender, rateLimit)
	reporter.Start()
	//reporter := usecase.NewMetricLoopReporter(sender)
	collector := usecase.NewCollector(reporter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collector.Collect()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				collector.Report()
				return
			case <-ticker.C:
				collector.Report()
			}
		}
	}()

	go func() {
		ticker := time.NewTicker(time.Duration(pollInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collector.CollectSystemMetrics()
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")
	reporter.Close()
	time.Sleep(100 * time.Millisecond)
	log.Println("Agent stopped")
}
