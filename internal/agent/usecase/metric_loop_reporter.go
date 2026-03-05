package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

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
