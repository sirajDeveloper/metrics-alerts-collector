package scheduler

import (
	"context"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/usecase"
)

type MetricEmitterScheduler struct {
	emitter  *usecase.MetricsEmitterService
	interval int
	restore  bool
}

func NewMetricEmitterScheduler(emitter *usecase.MetricsEmitterService, interval int, restore bool) *MetricEmitterScheduler {
	return &MetricEmitterScheduler{emitter: emitter, interval: interval, restore: restore}
}

func (s *MetricEmitterScheduler) Start(ctx context.Context) {
	if s.restore {
		s.emitter.EmitAll()
	}
	if s.interval <= 0 {
		return
	}
	ticker := time.NewTicker(time.Duration(s.interval) * time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.emitter.EmitAll()
			}
		}
	}()
}
