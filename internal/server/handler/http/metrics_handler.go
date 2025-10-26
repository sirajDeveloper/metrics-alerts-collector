package http

import (
	"embed"
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

//go:embed templates/*.html
var templatesFS embed.FS

type MetricsHandler struct {
	metricUpdater usecase.MetricUpdater
	metricGetter  usecase.MetricGetter
}

func NewMetricsHandler(metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter) *MetricsHandler {
	return &MetricsHandler{
		metricUpdater: metricUpdater,
		metricGetter:  metricGetter,
	}
}

func (h *MetricsHandler) UpdateCounter(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricUpdateRequest

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		logger.Log.Debug("cannot decode request JSON body", zap.Error(err))
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	if req.Name == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	if err := h.metricUpdater.MetricUpdate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricUpdateRequest

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	if req.Name == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	if err := h.metricUpdater.MetricUpdate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricValueRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	resp, err := h.metricGetter.GetMetricValue(&req)
	if err != nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricUpdateRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	if req.Type != "counter" && req.Type != "gauge" {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if req.Value == "" {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	if err := h.metricUpdater.MetricUpdate(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	displayMetrics := h.metricGetter.GetAllMetricsForDisplay()

	tmpl, err := template.ParseFS(templatesFS, "templates/metrics.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Metrics []dto.DisplayMetricDTO
	}{
		Metrics: displayMetrics,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
