package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

type mockMetricUpdater struct {
	updateFunc           func(req *dto.MetricUpdateRequest) error
	getFunc              func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error)
	getAllForDisplayFunc func() []dto.DisplayMetricDTO
}

func (m *mockMetricUpdater) MetricUpdate(req *dto.MetricUpdateRequest) error {
	if m.updateFunc != nil {
		return m.updateFunc(req)
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

func TestCounterHandler_UpdateCounter_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(req *dto.MetricUpdateRequest) error {
			if req.Type != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", req.Type)
			}
			if req.Name != "testCounter" {
				t.Errorf("expected metricName 'testCounter', got '%s'", req.Name)
			}
			if req.Value != "100" {
				t.Errorf("expected metricValue '100', got '%s'", req.Value)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "testCounter",
		Type:  "counter",
		Value: "100",
	}
	jsonBody, _ := json.Marshal(requestBody)

	r := chi.NewRouter()
	r.Post("/update/counter", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_MethodNotAllowed(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/counter", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/update/counter", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_EmptyMetricName(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "",
		Type:  "counter",
		Value: "100",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/counter", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_EmptyMetricValue(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "testCounter",
		Type:  "counter",
		Value: "",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/counter", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_UpdaterError(t *testing.T) {
	expectedError := errors.New("invalid counter value")

	mock := &mockMetricUpdater{
		updateFunc: func(req *dto.MetricUpdateRequest) error {
			return expectedError
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "testCounter",
		Type:  "counter",
		Value: "invalid",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/counter", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
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

func TestGaugeHandler_UpdateGauge_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(req *dto.MetricUpdateRequest) error {
			if req.Type != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", req.Type)
			}
			if req.Name != "testGauge" {
				t.Errorf("expected metricName 'testGauge', got '%s'", req.Name)
			}
			if req.Value != "123.45" {
				t.Errorf("expected metricValue '123.45', got '%s'", req.Value)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "testGauge",
		Type:  "gauge",
		Value: "123.45",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/gauge", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_MethodNotAllowed(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/gauge", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/update/gauge", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_EmptyMetricName(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "",
		Type:  "gauge",
		Value: "123.45",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/gauge", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_EmptyMetricValue(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "testGauge",
		Type:  "gauge",
		Value: "",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/gauge", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_UpdaterError(t *testing.T) {
	expectedError := errors.New("invalid gauge value")

	mock := &mockMetricUpdater{
		updateFunc: func(req *dto.MetricUpdateRequest) error {
			return expectedError
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "testGauge",
		Type:  "gauge",
		Value: "invalid",
	}
	jsonBody, _ := json.Marshal(requestBody)

	router := chi.NewRouter()
	router.Post("/update/gauge", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetMetricValue_Success_Counter(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
			if req.Type != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", req.Type)
			}
			if req.Name != "testCounter" {
				t.Errorf("expected metricName 'testCounter', got '%s'", req.Name)
			}
			return &dto.MetricValueResponse{
				Name:  "testCounter",
				Type:  "counter",
				Value: "100",
			}, nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricValueRequest{
		Name: "testCounter",
		Type: "counter",
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

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}

	var response dto.MetricValueResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Errorf("failed to unmarshal response: %v", err)
	}

	if response.Name != "testCounter" {
		t.Errorf("expected name 'testCounter', got '%s'", response.Name)
	}
	if response.Type != "counter" {
		t.Errorf("expected type 'counter', got '%s'", response.Type)
	}
	if response.Value != "100" {
		t.Errorf("expected value '100', got '%v'", response.Value)
	}
}

func TestGetMetricValue_Success_Gauge(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(req *dto.MetricValueRequest) (*dto.MetricValueResponse, error) {
			if req.Type != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", req.Type)
			}
			if req.Name != "testGauge" {
				t.Errorf("expected metricName 'testGauge', got '%s'", req.Name)
			}
			return &dto.MetricValueResponse{
				Name:  "testGauge",
				Type:  "gauge",
				Value: "123.45",
			}, nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricValueRequest{
		Name: "testGauge",
		Type: "gauge",
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

	if response.Name != "testGauge" {
		t.Errorf("expected name 'testGauge', got '%s'", response.Name)
	}
	if response.Type != "gauge" {
		t.Errorf("expected type 'gauge', got '%s'", response.Type)
	}
	if response.Value != "123.45" {
		t.Errorf("expected value '123.45', got '%v'", response.Value)
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
		Name: "unknown",
		Type: "counter",
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
		Name:  "testMetric",
		Type:  "unknown",
		Value: "100",
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
			if req.Type != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", req.Type)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "test",
		Type:  "counter",
		Value: "100",
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
			if req.Type != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", req.Type)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	requestBody := dto.MetricUpdateRequest{
		Name:  "test",
		Type:  "gauge",
		Value: "123.45",
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
