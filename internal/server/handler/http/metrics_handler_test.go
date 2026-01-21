package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

func TestMain(m *testing.M) {
	logger.InitLogger(true)
	defer func() {
		err := logger.Sync()
		if err != nil {
			logger.Log.Error("Error while logger.Sync")
		}
	}()

	os.Exit(m.Run())
}

type mockMetricUpdater struct {
	updateFunc           func(req *dto.MetricUpdateRequest) error
	getFunc              func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error)
	getAllForDisplayFunc func() []dto.DisplayMetricDTO
	getAllMetricsFunc    func() []*model.Metrics
}

var _ usecase.MetricGetter = (*mockMetricUpdater)(nil)

func (m *mockMetricUpdater) MetricUpdate(req *dto.MetricUpdateRequest) error {
	if m.updateFunc != nil {
		return m.updateFunc(req)
	}
	return nil
}

func (m *mockMetricUpdater) GetAllMetrics() []*model.Metrics {
	if m.getAllMetricsFunc != nil {
		return m.getAllMetricsFunc()
	}
	return nil
}

func (m *mockMetricUpdater) GetMetricValue(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
	if m.getFunc != nil {
		return m.getFunc(req)
	}
	return &dto.MetricValueResponse{}, nil
}

func (m *mockMetricUpdater) GetAllMetricsForDisplay() []dto.DisplayMetricDTO {
	if m.getAllForDisplayFunc != nil {
		return m.getAllForDisplayFunc()
	}
	return nil
}

func TestNewMetricsHandler(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	if handler == nil {
		t.Fatal("expected non-nil handler")
		return
	}

	if handler.metricUpdater != mock {
		t.Error("expected metricUpdater to be set correctly")
	}
}

func TestGetMetricValue_Success_Counter(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
			if req.MType != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", req.MType)
			}
			if req.ID != "testCounter" {
				t.Errorf("expected metricName 'testCounter', got '%s'", req.ID)
			}
			delta := int64(100)
			return &dto.MetricValueResponse{
				ID:    "testCounter",
				MType: "counter",
				Delta: &delta,
			}, nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricValueRequest{
		ID:    "testCounter",
		MType: "counter",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/value", handler.GetMetricValue)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", w.Header().Get("Content-Type"))
	}

	var response dto.MetricValueResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response.ID != "testCounter" {
		t.Errorf("expected ID 'testCounter', got '%s'", response.ID)
	}
	if response.MType != "counter" {
		t.Errorf("expected MType 'counter', got '%s'", response.MType)
	}
	if response.Delta == nil || *response.Delta != 100 {
		t.Errorf("expected Delta 100, got '%v'", response.Delta)
	}
}

func TestGetMetricValue_Success_Gauge(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
			if req.MType != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", req.MType)
			}
			if req.ID != "testGauge" {
				t.Errorf("expected metricName 'testGauge', got '%s'", req.ID)
			}
			value := 123.45
			return &dto.MetricValueResponse{
				ID:    "testGauge",
				MType: "gauge",
				Value: &value,
			}, nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricValueRequest{
		ID:    "testGauge",
		MType: "gauge",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/value", handler.GetMetricValue)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response dto.MetricValueResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response.ID != "testGauge" {
		t.Errorf("expected ID 'testGauge', got '%s'", response.ID)
	}
	if response.MType != "gauge" {
		t.Errorf("expected MType 'gauge', got '%s'", response.MType)
	}
	if response.Value == nil || *response.Value != 123.45 {
		t.Errorf("expected Value 123.45, got '%v'", response.Value)
	}
}

func TestGetMetricValue_NotFound(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
			return nil, errors.New("metric not found")
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricValueRequest{
		ID:    "unknown",
		MType: "counter",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/value", handler.GetMetricValue)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetAllMetrics_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		getAllForDisplayFunc: func() []dto.DisplayMetricDTO {
			return []dto.DisplayMetricDTO{
				{ID: "testGauge", MType: "gauge", ValueStr: "123.45"},
				{ID: "testCounter", MType: "counter", ValueStr: "100"},
			}
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Get("/", handler.GetAllMetrics)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected Content-Type 'text/html; charset=utf-8', got '%s'", w.Header().Get("Content-Type"))
	}

	body := w.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}

	if !strings.Contains(body, "testGauge") {
		t.Error("expected body to contain 'testGauge'")
	}
	if !strings.Contains(body, "testCounter") {
		t.Error("expected body to contain 'testCounter'")
	}
}

func TestGetAllMetrics_Empty(t *testing.T) {
	mock := &mockMetricUpdater{
		getAllForDisplayFunc: func() []dto.DisplayMetricDTO {
			return []dto.DisplayMetricDTO{}
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Get("/", handler.GetAllMetrics)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateMetric_UnknownType(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		ID:    "testMetric",
		MType: "unknown",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update", handler.UpdateMetric)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateMetric_Counter_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(req *dto.MetricUpdateRequest) error {
			if req.MType != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", req.MType)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	delta := int64(100)
	requestBody := dto.MetricUpdateRequest{
		ID:    "test",
		MType: "counter",
		Delta: &delta,
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update", handler.UpdateMetric)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateMetric_Gauge_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(req *dto.MetricUpdateRequest) error {
			if req.MType != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", req.MType)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	value := 123.45
	requestBody := dto.MetricUpdateRequest{
		ID:    "test",
		MType: "gauge",
		Value: &value,
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update", handler.UpdateMetric)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
