package main

import (
	"fmt"
	"log"
	"net/http"

	httpHandler "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

func main() {
	metricRepo := cache.NewMemStorage()
	metricService := usecase.NewMetricService(metricRepo)
	router := httpHandler.NewRouter(metricService)

	mux := http.NewServeMux()
	router.RegisterRoutes(mux)

	recoveryMux := recoverMiddleware(mux)

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", recoveryMux); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Println("Panic recovered:", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
