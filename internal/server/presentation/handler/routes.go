package handler

import (
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
}
