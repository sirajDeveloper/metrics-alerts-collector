package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"

	httpHandler "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/handler/http"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

func ExampleMetricsHandler_UpdateMetric() {
	repo := cache.NewMemStorage()
	service := usecase.NewMetricService(repo, nil)
	handler := httpHandler.NewMetricsHandler(service, service, nil)

	reqBody := dto.MetricUpdateRequest{
		ID:    "temperature",
		MType: "gauge",
	}
	value := 25.5
	reqBody.Value = &value

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateMetric(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))

	// Output:
	// Status: 200
	// Response:
}

func ExampleMetricsHandler_UpdateMetricURLParam() {
	repo := cache.NewMemStorage()
	service := usecase.NewMetricService(repo, nil)
	handler := httpHandler.NewMetricsHandler(service, service, nil)

	req := httptest.NewRequest("POST", "/update/gauge/temperature/25.5", nil)
	w := httptest.NewRecorder()

	handler.UpdateMetricURLParam(w, req)

	resp := w.Result()
	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 200
}

func ExampleMetricsHandler_UpdateMetrics() {
	repo := cache.NewMemStorage()
	service := usecase.NewMetricService(repo, nil)
	handler := httpHandler.NewMetricsHandler(service, service, nil)

	reqs := []dto.MetricUpdateRequest{
		{
			ID:    "cpu",
			MType: "gauge",
		},
		{
			ID:    "requests",
			MType: "counter",
		},
	}
	cpuValue := 45.2
	reqs[0].Value = &cpuValue
	delta := int64(10)
	reqs[1].Delta = &delta

	jsonData, _ := json.Marshal(reqs)
	req := httptest.NewRequest("POST", "/updates", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.UpdateMetrics(w, req)

	resp := w.Result()
	fmt.Printf("Status: %d\n", resp.StatusCode)

	// Output:
	// Status: 200
}

func ExampleMetricsHandler_GetMetricValue() {
	repo := cache.NewMemStorage()
	service := usecase.NewMetricService(repo, nil)

	metric := dto.MetricUpdateRequest{
		ID:    "temperature",
		MType: "gauge",
	}
	value := 25.5
	metric.Value = &value
	_ = service.MetricUpdate(&metric)

	handler := httpHandler.NewMetricsHandler(service, service, nil)

	reqBody := dto.MetricValueRequest{
		ID:    "temperature",
		MType: "gauge",
	}

	jsonData, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/value", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.GetMetricValue(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	var metricResp dto.MetricValueResponse
	_ = json.Unmarshal(body, &metricResp)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("ID: %s\n", metricResp.ID)
	fmt.Printf("Type: %s\n", metricResp.MType)
	if metricResp.Value != nil {
		fmt.Printf("Value: %.1f\n", *metricResp.Value)
	}

	// Output:
	// Status: 200
	// ID: temperature
	// Type: gauge
	// Value: 25.5
}

func ExampleMetricsHandler_GetMetricValueURLParam() {
	repo := cache.NewMemStorage()
	service := usecase.NewMetricService(repo, nil)

	metric := dto.MetricUpdateRequest{
		ID:    "requests",
		MType: "counter",
	}
	delta := int64(100)
	metric.Delta = &delta
	_ = service.MetricUpdate(&metric)

	handler := httpHandler.NewMetricsHandler(service, service, nil)

	req := httptest.NewRequest("GET", "/value/counter/requests", nil)
	w := httptest.NewRecorder()

	handler.GetMetricValueURLParam(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Value: %s\n", string(body))

	// Output:
	// Status: 200
	// Value: 100
}

func ExampleMetricsHandler_GetAllMetrics() {
	repo := cache.NewMemStorage()
	service := usecase.NewMetricService(repo, nil)

	metrics := []dto.MetricUpdateRequest{
		{ID: "cpu", MType: "gauge"},
		{ID: "memory", MType: "gauge"},
		{ID: "requests", MType: "counter"},
	}
	cpuVal := 45.2
	memVal := 78.5
	metrics[0].Value = &cpuVal
	metrics[1].Value = &memVal
	delta := int64(100)
	metrics[2].Delta = &delta

	for _, m := range metrics {
		_ = service.MetricUpdate(&m)
	}

	handler := httpHandler.NewMetricsHandler(service, service, nil)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.GetAllMetrics(w, req)

	resp := w.Result()

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))

	// Output:
	// Status: 200
	// Content-Type: text/html; charset=utf-8
}

func ExampleHealthHandler_Ping() {
	handler := httpHandler.NewHealthHandler(nil)

	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()

	handler.Ping(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Printf("Status: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))

	// Output:
	// Status: 500
	// Response: Internal Server Error
}
