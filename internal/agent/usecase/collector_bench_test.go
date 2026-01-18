package usecase

import (
	"testing"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type mockReporter struct{}

func (m *mockReporter) MetricsReport(metrics []domain.Metric) {}

func BenchmarkCollector_Collect(b *testing.B) {
	reporter := &mockReporter{}
	collector := NewCollector(reporter)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Collect()
	}
}

func BenchmarkCollector_CollectSystemMetrics(b *testing.B) {
	reporter := &mockReporter{}
	collector := NewCollector(reporter)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.CollectSystemMetrics()
	}
}

func BenchmarkCollector_GetMetrics(b *testing.B) {
	reporter := &mockReporter{}
	collector := NewCollector(reporter)

	for i := 0; i < 100; i++ {
		collector.Collect()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.GetMetrics()
	}
}

func BenchmarkCollector_Report(b *testing.B) {
	reporter := &mockReporter{}
	collector := NewCollector(reporter)

	for i := 0; i < 100; i++ {
		collector.Collect()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.Report()
	}
}
