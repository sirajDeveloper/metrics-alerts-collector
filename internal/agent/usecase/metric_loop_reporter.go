package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type MetricSender interface {
	Send(metric *domain.Metric)
}

type MetricLoopReporter struct {
	sender MetricSender
}

func NewMetricLoopReporter(sender MetricSender) *MetricLoopReporter {
	return &MetricLoopReporter{sender: sender}
}

var _ MetricsReporter = (*MetricLoopReporter)(nil)

func (r *MetricLoopReporter) MetricsReport(metrics []domain.Metric) {
	for _, m := range metrics {
		r.sender.Send(&m)
	}
}
