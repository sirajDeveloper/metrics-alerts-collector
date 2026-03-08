package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/infrastructure"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/buildinfo"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {

	buildinfo.PrintBuildInfo()
	ParseConfig()
	logger.InitLogger(false)

	/*var publicKey *rsa.PublicKey
	if cryptoKeyPath != "" {
		key, err := crypto.LoadPublicKey(cryptoKeyPath)
		if err != nil {
			log.Fatalf("Failed to load public key: %v", err)
		}
		publicKey = key
		log.Printf("Public key loaded from: %s", cryptoKeyPath)
	}*/
	serverURL := "http://" + address
	conn, err := grpc.NewClient(grpcAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("ошибка при установлении соединения с сервером %s", err)
	}
	defer conn.Close()
	c := proto.NewMetricsClient(conn)
	sender := infrastructure.NewMetricsClient(c, countRetrySave, 5*time.Second, agentIP)
	//sender := infrastructure.NewHTTPSender(serverURL, secretKey, countRetrySave, publicKey)
	fmt.Printf("HTTPSender init with serverURL: %v\n", serverURL)
	fmt.Printf("Rate limit: %d concurrent requests\n", rateLimit)
	reporter := usecase.NewMetricBatchReporter(sender)
	collector := usecase.NewCollector(reporter)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

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

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(time.Duration(reportInterval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				collector.Report()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

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
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	<-sigChan
	log.Println("Shutting down gracefully...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-shutdownCtx.Done():
		log.Println("Timeout waiting for background routines to finish")
	}

	//reporter.Close()
	log.Println("Agent stopped")
}
