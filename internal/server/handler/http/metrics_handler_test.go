package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

type mockMetricUpdater struct {
	updateFunc           func(metricType, metricName, metricValue string) error
	getFunc              func(metricType, metricName string) (string, error)
	getAllForDisplayFunc func() []dto.MetricDTO
}

func (m *mockMetricUpdater) MetricUpdate(metricType, metricName, metricValue string) error {
	if m.updateFunc != nil {
		return m.updateFunc(metricType, metricName, metricValue)
	}
	return nil
}

func (m *mockMetricUpdater) GetMetricValue(metricType, metricName string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(metricType, metricName)
	}
	return "", nil
}

func (m *mockMetricUpdater) GetAllMetricsForDisplay() []dto.MetricDTO {
	if m.getAllForDisplayFunc != nil {
		return m.getAllForDisplayFunc()
	}
	return nil
}

func TestCounterHandler_UpdateCounter_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			if metricType != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", metricType)
			}
			if metricName != "testCounter" {
				t.Errorf("expected metricName 'testCounter', got '%s'", metricName)
			}
			if metricValue != "100" {
				t.Errorf("expected metricValue '100', got '%s'", metricValue)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	r := chi.NewRouter()
	r.Post("/update/counter/{name}/{value}", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/100", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_MethodNotAllowed(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/update/counter/testCounter/100", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_EmptyMetricName(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter//100", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_EmptyMetricValue(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_UpdaterError(t *testing.T) {
	expectedError := errors.New("invalid counter value")

	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			return expectedError
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/counter/{name}/{value}", handler.UpdateCounter)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/invalid", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewCounterHandler(t *testing.T) {
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
		updateFunc: func(metricType, metricName, metricValue string) error {
			if metricType != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", metricType)
			}
			if metricName != "testGauge" {
				t.Errorf("expected metricName 'testGauge', got '%s'", metricName)
			}
			if metricValue != "123.45" {
				t.Errorf("expected metricValue '123.45', got '%s'", metricValue)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/gauge/{name}/{value}", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/123.45", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_MethodNotAllowed(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/gauge/{name}/{value}", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/update/gauge/testGauge/123.45", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_EmptyMetricName(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/gauge/{name}/{value}", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge//123.45", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_EmptyMetricValue(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/gauge/{name}/{value}", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_UpdaterError(t *testing.T) {
	expectedError := errors.New("invalid gauge value")

	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			return expectedError
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/gauge/{name}/{value}", handler.UpdateGauge)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/invalid", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetMetricValue_Success_Counter(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(metricType, metricName string) (string, error) {
			if metricType != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", metricType)
			}
			if metricName != "testCounter" {
				t.Errorf("expected metricName 'testCounter', got '%s'", metricName)
			}
			return "100", nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", handler.GetMetricValue)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/value/counter/testCounter", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "100" {
		t.Errorf("expected body '100', got '%s'", w.Body.String())
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got '%s'", w.Header().Get("Content-Type"))
	}
}

func TestGetMetricValue_Success_Gauge(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(metricType, metricName string) (string, error) {
			if metricType != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", metricType)
			}
			if metricName != "testGauge" {
				t.Errorf("expected metricName 'testGauge', got '%s'", metricName)
			}
			return "123.45", nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", handler.GetMetricValue)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/testGauge", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "123.45" {
		t.Errorf("expected body '123.45', got '%s'", w.Body.String())
	}
}

func TestGetMetricValue_NotFound(t *testing.T) {
	mock := &mockMetricUpdater{
		getFunc: func(metricType, metricName string) (string, error) {
			return "", errors.New("metric not found")
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Get("/value/{type}/{name}", handler.GetMetricValue)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/value/counter/unknown", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetAllMetrics_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		getAllForDisplayFunc: func() []dto.MetricDTO {
			return []dto.MetricDTO{
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
}

func TestGetAllMetrics_Empty(t *testing.T) {
	mock := &mockMetricUpdater{
		getAllForDisplayFunc: func() []dto.MetricDTO {
			return []dto.MetricDTO{}
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

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/unknown/testMetric/100", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateMetric_Counter_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			if metricType != "counter" {
				t.Errorf("expected metricType 'counter', got '%s'", metricType)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/counter/test/100", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestUpdateMetric_Gauge_Success(t *testing.T) {
	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			if metricType != "gauge" {
				t.Errorf("expected metricType 'gauge', got '%s'", metricType)
			}
			return nil
		},
	}

	handler := NewMetricsHandler(mock, mock)

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.UpdateMetric)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123.45", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}
