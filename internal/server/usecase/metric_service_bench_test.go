package usecase

import (
	"testing"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/datastorage/cache"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase/dto"
)

func BenchmarkMetricService_MetricUpdate(b *testing.B) {
	repo := cache.NewMemStorage()
	service := NewMetricService(repo, nil, nil)

	req := &dto.MetricUpdateRequest{
		ID:    "testMetric",
		MType: "gauge",
	}
	val := 123.45
	req.Value = &val

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.MetricUpdate(req)
	}
}

func BenchmarkMetricService_GetMetricValue(b *testing.B) {
	repo := cache.NewMemStorage()
	service := NewMetricService(repo, nil, nil)

	metric := model.CreateMetric("testMetric", "gauge")
	val := 123.45
	metric.Value = &val
	repo.Save(metric)

	req := &dto.MetricValueRequest{
		ID:    "testMetric",
		MType: "gauge",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.GetMetricValue(req)
	}
}

func BenchmarkMetricService_GetAllMetricsForDisplay(b *testing.B) {
	repo := cache.NewMemStorage()
	service := NewMetricService(repo, nil, nil)

	for i := 0; i < 1000; i++ {
		metric := model.CreateMetric("testMetric", "gauge")
		val := float64(i)
		metric.Value = &val
		repo.Save(metric)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.GetAllMetricsForDisplay()
	}
}
