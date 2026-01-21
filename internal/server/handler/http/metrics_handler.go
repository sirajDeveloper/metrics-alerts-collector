package http

import (
	"embed"
	"encoding/json"
	"errors"
	"strconv"

	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
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

func (h *MetricsHandler) GetMetricValueURLParam(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	req := dto.MetricValueRequest{
		ID: metricName, MType: metricType,
	}

	value, err := h.metricGetter.GetMetricValue(&req)
	if err != nil {
		logger.Log.Error("GetMetricValueURLParam not found", zap.Any("request", req), zap.Error(err))
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	var valueStr string
	switch value.MType {
	case "gauge":
		if value.Value != nil {
			valueStr = strconv.FormatFloat(*value.Value, 'f', -1, 64)
		} else {
			logger.Log.Error("GetMetricValueURLParam gauge nil", zap.Any("value", value))
			http.Error(w, "Gauge value is nil", http.StatusInternalServerError)
			return
		}
	case "counter":
		if value.Delta != nil {
			valueStr = strconv.FormatInt(*value.Delta, 10)
		} else {
			logger.Log.Error("GetMetricValueURLParam counter nil", zap.Any("value", value))
			http.Error(w, "Counter value is nil", http.StatusInternalServerError)
			return
		}
	default:
		logger.Log.Error("GetMetricValueURLParam invalid type", zap.Any("value", value))
		http.Error(w, "Invalid metric type", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(valueStr))
	if err != nil {
		logger.Log.Error("GetMetricValueURLParam write error", zap.Error(err))
		http.Error(w, "Error while response writing", http.StatusInternalServerError)
		return
	}
}

func (h *MetricsHandler) UpdateMetricURLParam(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricType != "counter" && metricType != "gauge" {
		logger.Log.Error("UpdateMetricURLParam unknown type", zap.String("type", metricType))
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return
	}

	if metricName == "" {
		logger.Log.Error("UpdateMetricURLParam name required")
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return
	}

	if metricValue == "" {
		logger.Log.Error("UpdateMetricURLParam value invalid")
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
			logger.Log.Error("UpdateMetricURLParam invalid gauge", zap.String("value", metricValue), zap.Error(err))
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		req.Value = &val
	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			logger.Log.Error("UpdateMetricURLParam invalid counter", zap.String("value", metricValue), zap.Error(err))
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		req.Delta = &val
	}
	if err := h.metricUpdater.MetricUpdate(&req); err != nil {
		logger.Log.Error("UpdateMetricURLParam update error", zap.Any("request", req), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricValueRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Error("GetMetricValue decode error", zap.Error(err))
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	resp, err := h.metricGetter.GetMetricValue(&req)
	if err != nil {
		logger.Log.Error("GetMetricValue not found", zap.Any("request", req), zap.Error(err))
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	if logger.Log != nil {
		logger.Log.Info("GetMetricValue", zap.Any("requestBody", req), zap.Any("responseBody", resp))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&resp); err != nil {
		logger.Log.Error("GetMetricValue encode error", zap.Any("response", resp), zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricUpdateRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Error("UpdateMetric decode error", zap.Error(err))
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	if err := h.processMetricUpdate(&req, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	var reqs []dto.MetricUpdateRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqs); err != nil {
		logger.Log.Error("UpdateMetrics decode error", zap.Error(err))
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	for i := range reqs {
		if err := h.processMetricUpdate(&reqs[i], w); err != nil {
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) processMetricUpdate(req *dto.MetricUpdateRequest, w http.ResponseWriter) error {
	if req.MType != "counter" && req.MType != "gauge" {
		logger.Log.Error("UpdateMetric unknown type", zap.String("type", req.MType))
		http.Error(w, "Unknown metric type", http.StatusBadRequest)
		return errors.New("unknown metric type")
	}

	if req.ID == "" {
		logger.Log.Error("UpdateMetric name required")
		http.Error(w, "Metric name is required", http.StatusBadRequest)
		return errors.New("metric name is required")
	}

	if req.MType == "gauge" && req.Value == nil {
		logger.Log.Error("UpdateMetric gauge required")
		http.Error(w, "Gauge value is required", http.StatusBadRequest)
		return errors.New("gauge value is required")
	}

	if req.MType == "counter" && req.Delta == nil {
		logger.Log.Error("UpdateMetric counter required")
		http.Error(w, "Counter delta is required", http.StatusBadRequest)
		return errors.New("counter delta is required")
	}

	if logger.Log != nil {
		logger.Log.Info("UpdateMetric", zap.Any("requestBody", req))
	}

	if err := h.metricUpdater.MetricUpdate(req); err != nil {
		logger.Log.Error("UpdateMetric update error", zap.Any("request", req), zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return err
	}

	return nil
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
		logger.Log.Error("GetAllMetrics template error", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
