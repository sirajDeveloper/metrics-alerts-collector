package cache

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type metricKey struct {
	ID    string
	MType string
}

type MemStorage struct {
	cache map[metricKey]*model.Metrics
}

var _ repository.MetricRepository = (*MemStorage)(nil)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		cache: make(map[metricKey]*model.Metrics),
	}
}

func (m *MemStorage) GetMetric(mType, name string) *model.Metrics {
	key := metricKey{ID: name, MType: mType}
	return m.cache[key]
}

func (m *MemStorage) Save(metrics *model.Metrics) {
	key := metricKey{ID: metrics.ID, MType: metrics.MType}
	m.cache[key] = metrics
}

func (m *MemStorage) GetAll() []*model.Metrics {
	result := make([]*model.Metrics, 0, len(m.cache))
	for _, metric := range m.cache {
		result = append(result, metric)
	}
	return result
}
