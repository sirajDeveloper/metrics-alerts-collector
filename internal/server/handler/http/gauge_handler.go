package http

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/application"
	"net/http"
	"strings"
)

type GaugeHandler struct {
	metricUpdater application.MetricUpdater
}

func NewGaugeHandler(metricUpdater application.MetricUpdater) *GaugeHandler {
	return &GaugeHandler{
		metricUpdater: metricUpdater,
	}
}

func (h *GaugeHandler) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/update/gauge/")
	parts := strings.Split(path, "/")

	if len(parts) != 2 {
		http.Error(w, "Invalid path format", http.StatusNotFound)
		return
	}

	metricName := parts[0]
	metricValue := parts[1]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	if metricValue == "" {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	if err := h.metricUpdater.MetricUpdate("gauge", metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
