package cache

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type MemStorage struct {
	cache map[model.Metrics]struct{}
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		cache: make(map[model.Metrics]struct{}),
	}
}

var _ repository.MetricRepository = (*MemStorage)(nil)

func (m *MemStorage) GetMetric(mType, name string) *model.Metrics {
	for metrics := range m.cache {
		if metrics.MType == mType && metrics.ID == name {
			return &metrics
		}
	}
	return nil
}

func (m *MemStorage) Save(metrics *model.Metrics) {
	m.cache[*metrics] = struct{}{}
}
