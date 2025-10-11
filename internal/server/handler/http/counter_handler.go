package http

import (
	"net/http"
	"strings"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type CounterHandler struct {
	metricUpdater usecase.MetricUpdater
}

func NewCounterHandler(metricUpdater usecase.MetricUpdater) *CounterHandler {
	return &CounterHandler{
		metricUpdater: metricUpdater,
	}
}

func (h *CounterHandler) UpdateCounter(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/update/counter/")
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

	if err := h.metricUpdater.MetricUpdate("counter", metricName, metricValue); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
