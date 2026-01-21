package usecase

import (
	"sync"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type MetricWorkerPoolReporter struct {
	sender      MetricSender
	metricsChan chan domain.Metric
	workerCount int
	wg          sync.WaitGroup
}

func NewMetricWorkerPoolReporter(sender MetricSender, workerCount int) *MetricWorkerPoolReporter {
	reporter := &MetricWorkerPoolReporter{
		sender:      sender,
		metricsChan: make(chan domain.Metric, workerCount*2),
		workerCount: workerCount,
	}

	return reporter
}

func (r *MetricWorkerPoolReporter) Start() {
	for i := 0; i < r.workerCount; i++ {
		r.wg.Add(1)
		go func() {
			defer r.wg.Done()
			for metric := range r.metricsChan {
				r.sender.Send(&metric)
			}
		}()
	}
}

var _ MetricsReporter = (*MetricWorkerPoolReporter)(nil)

func (r *MetricWorkerPoolReporter) MetricsReport(metrics []domain.Metric) {
	for _, m := range metrics {
		r.metricsChan <- m
	}
}

func (r *MetricWorkerPoolReporter) Close() {
	close(r.metricsChan)
	r.wg.Wait()
}
