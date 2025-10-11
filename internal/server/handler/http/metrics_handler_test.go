package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockMetricUpdater struct {
	updateFunc func(metricType, metricName, metricValue string) error
}

func (m *mockMetricUpdater) MetricUpdate(metricType, metricName, metricValue string) error {
	if m.updateFunc != nil {
		return m.updateFunc(metricType, metricName, metricValue)
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

	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/100", nil)

	handler.UpdateCounter(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_MethodNotAllowed(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/update/counter/testCounter/100", nil)

	handler.UpdateCounter(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_InvalidPathFormat(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "only metric name without value",
			path: "/update/counter/testCounter",
		},
		{
			name: "empty path",
			path: "/update/counter/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mock := &mockMetricUpdater{}
			handler := NewMetricsHandler(mock)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodPost, tt.path, nil)

			handler.UpdateCounter(w, r)

			if w.Code != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
			}
		})
	}
}

func TestCounterHandler_UpdateCounter_EmptyMetricName(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/counter//100", nil)

	handler.UpdateCounter(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_EmptyMetricValue(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/", nil)

	handler.UpdateCounter(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestCounterHandler_UpdateCounter_UpdaterError(t *testing.T) {
	expectedError := errors.New("invalid counter value")

	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			return expectedError
		},
	}

	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/counter/testCounter/invalid", nil)

	handler.UpdateCounter(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestNewCounterHandler(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

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

	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/123.45", nil)

	handler.UpdateGauge(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_MethodNotAllowed(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/update/gauge/testGauge/123.45", nil)

	handler.UpdateGauge(w, r)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_InvalidPathFormat(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "only metric name without value",
			path: "/update/gauge/testGauge",
		},
		{
			name: "empty path",
			path: "/update/gauge/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mock := &mockMetricUpdater{}
			handler := NewMetricsHandler(mock)

			w := httptest.NewRecorder()

			r := httptest.NewRequest(http.MethodPost, tt.path, nil)

			handler.UpdateGauge(w, r)

			if w.Code != http.StatusNotFound {
				t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
			}
		})
	}
}

func TestGaugeHandler_UpdateGauge_EmptyMetricName(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/gauge//123.45", nil)

	handler.UpdateGauge(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_EmptyMetricValue(t *testing.T) {
	mock := &mockMetricUpdater{}
	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/", nil)

	handler.UpdateGauge(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGaugeHandler_UpdateGauge_UpdaterError(t *testing.T) {
	expectedError := errors.New("invalid gauge value")

	mock := &mockMetricUpdater{
		updateFunc: func(metricType, metricName, metricValue string) error {
			return expectedError
		},
	}

	handler := NewMetricsHandler(mock)

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/update/gauge/testGauge/invalid", nil)

	handler.UpdateGauge(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
