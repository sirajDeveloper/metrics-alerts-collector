package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

func main() {
	parseConfig()

	metricRepo := cache.NewMemStorage()
	metricService := usecase.NewMetricService(metricRepo)
	chiRouter := router.NewChiRouter(metricService, metricService)

	server := &http.Server{
		Addr:    address,
		Handler: chiRouter.Handler(),
	}

	go func() {
		log.Println("Server starting on http://" + address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Server failed to start:", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutting down gracefully...")

	ctxt, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := server.Shutdown(ctxt); err != nil {
		log.Fatal()
	}
	time.Sleep(100 * time.Millisecond)
	log.Println("Server stopped")
}
