package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type MetricWorkerPoolReporter struct {
	sender      MetricSender
	metricsChan chan domain.Metric
	workerCount int
}

func NewMetricWorkerPoolReporter(sender MetricSender, workerCount int) *MetricWorkerPoolReporter {
	reporter := &MetricWorkerPoolReporter{
		sender:      sender,
		metricsChan: make(chan domain.Metric, workerCount*2),
		workerCount: workerCount,
	}

	for i := 0; i < workerCount; i++ {
		go func() {
			for metric := range reporter.metricsChan {
				reporter.sender.Send(&metric)
			}
		}()
	}

	return reporter
}

var _ MetricsReporter = (*MetricWorkerPoolReporter)(nil)

func (r *MetricWorkerPoolReporter) MetricsReport(metrics []domain.Metric) {
	for _, m := range metrics {
		r.metricsChan <- m
	}
}

func (r *MetricWorkerPoolReporter) Close() {
	close(r.metricsChan)
}
