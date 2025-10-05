package handler

import (
	"net/http"
	"strings"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/application"
)

type MetricsHandler struct {
	metricUpdater application.MetricUpdater
}

func NewMetricsHandler(metricUpdater application.MetricUpdater) *MetricsHandler {
	return &MetricsHandler{
		metricUpdater: metricUpdater,
	}
}

func (h *MetricsHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/update", h.UpdateMetric)
}

func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/update/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	metricType := parts[0]
	metricName := parts[1]
	metricValue := parts[2]

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	if err := h.metricUpdater.MetricUpdate(metricType, metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
