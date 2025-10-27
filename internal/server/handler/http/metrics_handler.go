package http

import (
	"embed"
	"encoding/json"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"
	"strconv"

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

func (h *MetricsHandler) GetMetricValueURLParam(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	req := dto.MetricValueRequest{
		ID: metricName, MType: metricType,
	}

	value, err := h.metricGetter.GetMetricValue(&req)
	if err != nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	var valueStr string
	switch value.MType {
	case "gauge":
		if value.Value != nil {
			valueStr = strconv.FormatFloat(*value.Value, 'f', -1, 64)
		} else {
			http.Error(w, "Gauge value is nil", http.StatusInternalServerError)
			return
		}
	case "counter":
		if value.Delta != nil {
			valueStr = strconv.FormatInt(*value.Delta, 10)
		} else {
			http.Error(w, "Counter value is nil", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Invalid metric type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(valueStr))
	if err != nil {
		http.Error(w, "Error while response writing", http.StatusInternalServerError)
		return
	}
}

func (h *MetricsHandler) UpdateMetricURLParam(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricType != "counter" && metricType != "gauge" {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if metricValue == "" {
		http.Error(w, "Invalid metric value", http.StatusBadRequest)
		return
	}

	var req dto.MetricUpdateRequest
	req.ID = metricName
	req.MType = metricType

	switch metricType {
	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		req.Value = &val
	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		req.Delta = &val
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

	logger.Log.Info("Payload", zap.Any("requestBody", req), zap.Any("responseBody", resp))

	w.Header().Set("Content-Type", "application/json")
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

	if req.MType != "counter" && req.MType != "gauge" {
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if req.MType == "gauge" && req.Value == nil {
		http.Error(w, "Gauge value is required", http.StatusBadRequest)
		return
	}

	if req.MType == "counter" && req.Delta == nil {
		http.Error(w, "Counter delta is required", http.StatusBadRequest)
		return
	}

	logger.Log.Info("Payload", zap.Any("requestBody", req))

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
