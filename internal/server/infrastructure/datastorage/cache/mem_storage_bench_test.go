package cache

import (
	"testing"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

func BenchmarkMemStorage_GetMetric(b *testing.B) {
	storage := NewMemStorage()
	metric := model.CreateMetric("testMetric", "gauge")
	val := 123.45
	metric.Value = &val
	storage.Save(metric)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetMetric("gauge", "testMetric")
	}
}

func BenchmarkMemStorage_Save(b *testing.B) {
	storage := NewMemStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metric := model.CreateMetric("testMetric", "gauge")
		val := float64(i)
		metric.Value = &val
		storage.Save(metric)
	}
}

func BenchmarkMemStorage_GetAll(b *testing.B) {
	storage := NewMemStorage()

	for i := 0; i < 1000; i++ {
		metric := model.CreateMetric("testMetric", "gauge")
		val := float64(i)
		metric.Value = &val
		storage.Save(metric)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = storage.GetAll()
	}
}
