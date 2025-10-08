package usecase

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type MetricSender interface {
	Send(metric domain.Metric) error
}
