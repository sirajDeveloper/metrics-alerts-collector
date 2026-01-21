package http

import (
	"embed"
	"encoding/json"
	"errors"
	"net"
	"strconv"
	"strings"

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

// MetricsHandler обрабатывает HTTP-запросы для работы с метриками.
// Поддерживает обновление и получение метрик через JSON и URL параметры.
// Эндпоинты:
//   - POST /update - обновление одной метрики (JSON)
//   - POST /update/{type}/{name}/{value} - обновление метрики через URL параметры
//   - POST /updates - массовое обновление метрик (JSON)
//   - POST /value - получение значения метрики (JSON)
//   - GET /value/{type}/{name} - получение значения метрики через URL параметры
//   - GET / - получение всех метрик в виде HTML страницы
type MetricsHandler struct {
	metricUpdater usecase.MetricUpdater
	metricGetter  usecase.MetricGetter
}

// NewMetricsHandler создает новый экземпляр MetricsHandler.
//
// Параметры:
//   - metricUpdater: интерфейс для обновления метрик
//   - metricGetter: интерфейс для получения метрик
//
// Возвращает новый экземпляр MetricsHandler.
func NewMetricsHandler(metricUpdater usecase.MetricUpdater, metricGetter usecase.MetricGetter) *MetricsHandler {
	return &MetricsHandler{
		metricUpdater: metricUpdater,
		metricGetter:  metricGetter,
	}
}

func (h *MetricsHandler) getIPAddress(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// GetMetricValueURLParam обрабатывает GET /value/{type}/{name}.
// Возвращает значение метрики в виде строки (text/plain).
//
// Параметры URL:
//   - type: тип метрики ("gauge" или "counter")
//   - name: имя метрики
//
// Ответы:
//   - 200 OK: значение метрики в виде строки
//   - 404 Not Found: метрика не найдена
//   - 500 Internal Server Error: внутренняя ошибка
//
// @Summary Получить значение метрики (URL)
// @Description Возвращает значение метрики в виде строки через URL параметры
// @Tags metrics
// @Produce text/plain
// @Param type path string true "Тип метрики" Enums(gauge, counter)
// @Param name path string true "Имя метрики"
// @Success 200 {string} string "Значение метрики"
// @Failure 404 "Метрика не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /value/{type}/{name} [get]
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

// UpdateMetricURLParam обрабатывает POST /update/{type}/{name}/{value}.
// Обновляет метрику через параметры URL.
//
// Параметры URL:
//   - type: тип метрики ("gauge" или "counter")
//   - name: имя метрики
//   - value: значение метрики (число в виде строки)
//
// Ответы:
//   - 200 OK: метрика успешно обновлена
//   - 400 Bad Request: некорректные параметры запроса
//
// @Summary Обновить метрику через URL
// @Description Обновляет метрику через параметры URL
// @Tags metrics
// @Param type path string true "Тип метрики" Enums(gauge, counter)
// @Param name path string true "Имя метрики"
// @Param value path string true "Значение метрики"
// @Success 200 "Метрика успешно обновлена"
// @Failure 400 "Некорректные параметры"
// @Router /update/{type}/{name}/{value} [post]
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
	req.IPAddress = h.getIPAddress(r)

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

// GetMetricValue обрабатывает POST /value.
// Возвращает значение метрики в формате JSON.
//
// Формат запроса (JSON):
//
//	{
//	  "id": "metricName",
//	  "type": "gauge"
//	}
//
// Формат ответа (JSON):
//
//	{
//	  "id": "metricName",
//	  "type": "gauge",
//	  "value": 123.45
//	}
//
// Ответы:
//   - 200 OK: успешное получение метрики
//   - 404 Not Found: метрика не найдена
//   - 500 Internal Server Error: ошибка при обработке запроса
//
// @Summary Получить значение метрики (JSON)
// @Description Возвращает значение метрики в формате JSON
// @Tags metrics
// @Accept json
// @Produce json
// @Param request body dto.MetricValueRequest true "Запрос метрики"
// @Success 200 {object} dto.MetricValueResponse "Значение метрики"
// @Failure 404 "Метрика не найдена"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /value [post]
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

// UpdateMetric обрабатывает POST /update.
// Обновляет или создает метрику из JSON-тела запроса.
//
// Формат запроса для gauge (JSON):
//
//	{
//	  "id": "metricName",
//	  "type": "gauge",
//	  "value": 123.45
//	}
//
// Формат запроса для counter (JSON):
//
//	{
//	  "id": "metricName",
//	  "type": "counter",
//	  "delta": 10
//	}
//
// Ответы:
//   - 200 OK: метрика успешно обновлена
//   - 400 Bad Request: некорректный формат запроса или данные
//   - 500 Internal Server Error: ошибка при обработке запроса
//
// @Summary Обновить метрику
// @Description Обновляет или создает метрику из JSON-тела запроса
// @Tags metrics
// @Accept json
// @Produce json
// @Param metric body dto.MetricUpdateRequest true "Данные метрики"
// @Success 200 "Метрика успешно обновлена"
// @Failure 400 "Некорректный формат запроса"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /update [post]
func (h *MetricsHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	var req dto.MetricUpdateRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Error("UpdateMetric decode error", zap.Error(err))
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	req.IPAddress = h.getIPAddress(r)

	if err := h.processMetricUpdate(&req, w); err != nil {
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UpdateMetrics обрабатывает POST /updates.
// Обновляет множество метрик из JSON-массива запросов.
//
// Формат запроса (JSON массив):
//
//	[
//	  {"id": "cpu", "type": "gauge", "value": 45.2},
//	  {"id": "memory", "type": "gauge", "value": 78.5},
//	  {"id": "requests", "type": "counter", "delta": 100}
//	]
//
// Ответы:
//   - 200 OK: все метрики успешно обновлены
//   - 400 Bad Request: некорректный формат запроса или данных
//   - 500 Internal Server Error: ошибка при обработке запроса
//
// @Summary Массовое обновление метрик
// @Description Обновляет множество метрик из JSON-массива запросов
// @Tags metrics
// @Accept json
// @Produce json
// @Param metrics body []dto.MetricUpdateRequest true "Массив метрик"
// @Success 200 "Все метрики успешно обновлены"
// @Failure 400 "Некорректный формат запроса"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router /updates [post]
func (h *MetricsHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	var reqs []dto.MetricUpdateRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&reqs); err != nil {
		logger.Log.Error("UpdateMetrics decode error", zap.Error(err))
		http.Error(w, "cannot decode request JSON body", http.StatusInternalServerError)
		return
	}

	ipAddress := h.getIPAddress(r)
	for i := range reqs {
		reqs[i].IPAddress = ipAddress
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

// GetAllMetrics обрабатывает GET /.
// Возвращает HTML-страницу со списком всех метрик.
//
// Ответы:
//   - 200 OK: HTML-страница со списком метрик
//   - 500 Internal Server Error: ошибка при генерации страницы
//
// @Summary Получить все метрики
// @Description Возвращает HTML-страницу со списком всех метрик
// @Tags metrics
// @Produce text/html
// @Success 200 {string} string "HTML страница с метриками"
// @Failure 500 "Внутренняя ошибка сервера"
// @Router / [get]
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
