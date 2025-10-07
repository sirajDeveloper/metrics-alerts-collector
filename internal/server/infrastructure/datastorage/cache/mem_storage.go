package cache

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type MemStorage struct {
	cache map[domain.Metrics]struct{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		cache: make(map[domain.Metrics]struct{}),
	}
}

var _ repository.MetricRepository = (*MemStorage)(nil)

func (m *MemStorage) GetMetric(mType, name string) *domain.Metrics {
	for metrics := range m.cache {
		if metrics.MType == mType && metrics.ID == name {
			return &metrics
		}
	}
	return nil
}

func (m *MemStorage) Save(metrics *domain.Metrics) {
	m.cache[*metrics] = struct{}{}
}
