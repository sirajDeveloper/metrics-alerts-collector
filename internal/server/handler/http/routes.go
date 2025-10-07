package http

import (
	"log"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/application"
)

type Router struct {
	counterHandler *CounterHandler
	gaugeHandler   *GaugeHandler
}

func NewRouter(metricUpdater application.MetricUpdater) *Router {
	return &Router{
		counterHandler: NewCounterHandler(metricUpdater),
		gaugeHandler:   NewGaugeHandler(metricUpdater),
	}
}

func (r *Router) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/update/counter/", r.counterHandler.UpdateCounter)
	mux.HandleFunc("/update/gauge/", r.gaugeHandler.UpdateGauge)
	mux.HandleFunc("/update/", r.handleUnknownUpdate)
}

func (r *Router) handleUnknownUpdate(w http.ResponseWriter, req *http.Request) {
	log.Printf("handleUnknownUpdate called for URL: %s", req.URL.Path)
	http.Error(w, "Unknown metric type", http.StatusBadRequest)
}
