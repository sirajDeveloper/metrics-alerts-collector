package infrastructure

import (
	"testing"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

func BenchmarkHTTPSender_prepareMetricRequest(b *testing.B) {
	sender := NewHTTPSender("http://localhost:8080", "", 1)

	gaugeMetric := &domain.Metric{
		Name:  "testGauge",
		Type:  domain.Gauge,
		Value: 123.45,
	}

	counterMetric := &domain.Metric{
		Name:  "testCounter",
		Type:  domain.Counter,
		Value: int64(100),
	}

	b.Run("Gauge", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = sender.prepareMetricRequest(gaugeMetric)
		}
	})

	b.Run("Counter", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = sender.prepareMetricRequest(counterMetric)
		}
	})
}

func BenchmarkCompressBody(b *testing.B) {
	data := []byte(`{"id":"testMetric","type":"gauge","value":123.45}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = compressBody(data, "application/json")
	}
}
