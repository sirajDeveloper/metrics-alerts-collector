package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type MetricBatchSender interface {
	SendBatch(metrics []domain.Metric)
}

type MetricBatchReporter struct {
	sender MetricBatchSender
}

func NewMetricBatchReporter(sender MetricBatchSender) *MetricBatchReporter {
	return &MetricBatchReporter{sender: sender}
}

var _ MetricsReporter = (*MetricBatchReporter)(nil)

func (r *MetricBatchReporter) MetricsReport(metrics []domain.Metric) {
	r.sender.SendBatch(metrics)
}
