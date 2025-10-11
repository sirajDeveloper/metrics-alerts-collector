package http

import (
	"embed"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
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
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
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

func (h *MetricsHandler) UpdateGauge(w http.ResponseWriter, r *http.Request) {
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
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

func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	value, err := h.metricGetter.GetMetricValue(metricType, metricName)
	if err != nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(value))
}

func (h *MetricsHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	displayMetrics := h.metricGetter.GetAllMetricsForDisplay()

	tmpl, err := template.ParseFS(templatesFS, "templates/metrics.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Metrics []dto.MetricDTO
	}{
		Metrics: displayMetrics,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
