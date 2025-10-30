package event

import "github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"

type MetricsEvent struct {
	Metrics *model.Metrics
}

type MetricsSender interface {
	Send(e MetricsEvent)
}
