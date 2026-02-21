package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/infrastructure"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/pkg/crypto"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	printBuildInfo()
	ParseConfig()
	logger.InitLogger(false)

	var publicKey *rsa.PublicKey
	if cryptoKeyPath != "" {
		key, err := crypto.LoadPublicKey(cryptoKeyPath)
		if err != nil {
			log.Fatalf("Failed to load public key: %v", err)
		}
		publicKey = key
		log.Printf("Public key loaded from: %s", cryptoKeyPath)
	}

	serverURL := "http://" + address
	sender := infrastructure.NewHTTPSender(serverURL, secretKey, countRetrySave, publicKey)
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

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}

	fmt.Fprintf(os.Stdout, "Build version: %s\n", version)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", date)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", commit)
}
