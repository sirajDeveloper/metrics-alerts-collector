package main

import (
	"log"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type Environments struct {
}

func main() {
	parseFlags()
	metricRepo := cache.NewMemStorage()
	metricService := usecase.NewMetricService(metricRepo)
	chiRouter := router.NewChiRouter(metricService, metricService)

	server := &http.Server{
		Addr:    address,
		Handler: chiRouter.Handler(),
	}

	log.Println("Server starting on http://" + address)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
