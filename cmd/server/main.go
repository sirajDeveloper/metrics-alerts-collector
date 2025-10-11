package main

import (
	"log"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/router"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

func main() {
	metricRepo := cache.NewMemStorage()
	metricService := usecase.NewMetricService(metricRepo)
	chiRouter := router.NewChiRouter(metricService, metricService)

	server := &http.Server{
		Addr:    ":8080",
		Handler: chiRouter.Handler(),
	}

	log.Println("Server starting on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
