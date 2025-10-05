package main

import (
	"log"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/application"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/repository"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/presentation/handler"
)

func main() {
	metricRepo := repository.NewMemStorage()
	metricService := application.NewMetricService(metricRepo)
	metricsHandler := handler.NewMetricsHandler(metricService)

	mux := http.NewServeMux()
	metricsHandler.RegisterRoutes(mux)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
