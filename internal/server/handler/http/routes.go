package http

import (
	"log"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type Router struct {
	metricsHandler *MetricsHandler
}

func NewRouter(metricUpdater usecase.MetricUpdater) *Router {
	return &Router{
		metricsHandler: NewMetricsHandler(metricUpdater),
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/update/counter/", r.metricsHandler.UpdateCounter)
	mux.HandleFunc("/update/gauge/", r.metricsHandler.UpdateGauge)
	mux.HandleFunc("/update/", r.handleUnknownUpdate)
}

func (r *Router) handleUnknownUpdate(w http.ResponseWriter, req *http.Request) {
	log.Printf("handleUnknownUpdate called for URL: %s", req.URL.Path)
	http.Error(w, "Unknown metric type", http.StatusBadRequest)
}
