package cache

import (
	"sync"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/repository"
)

type metricKey struct {
	ID    string
	MType string
}

type MemStorage struct {
	sync.RWMutex
	cache map[metricKey]*model.Metrics
}

var _ repository.MetricRepository = (*MemStorage)(nil)

func NewMemStorage() *MemStorage {
	return &MemStorage{
		cache: make(map[metricKey]*model.Metrics),
	}
}

func (m *MemStorage) GetMetric(mType, name string) *model.Metrics {
	m.RLock()
	defer m.RUnlock()
	key := metricKey{ID: name, MType: mType}
	metric := m.cache[key]
	if metric == nil {
		return nil
	}
	result := &model.Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}
	if metric.Delta != nil {
		result.Delta = metric.Delta
	}
	if metric.Value != nil {
		result.Value = metric.Value
	}
	return result
}

func (m *MemStorage) Save(metrics *model.Metrics) {
	m.Lock()
	defer m.Unlock()
	key := metricKey{ID: metrics.ID, MType: metrics.MType}
	m.cache[key] = metrics
}

func (m *MemStorage) GetAll() []*model.Metrics {
	m.RLock()
	defer m.RUnlock()
	if len(m.cache) == 0 {
		return nil
	}
	result := make([]*model.Metrics, 0, len(m.cache))
	for _, metric := range m.cache {
		if metric == nil {
			continue
		}
		copyM := &model.Metrics{
			ID:    metric.ID,
			MType: metric.MType,
		}
		if metric.Delta != nil {
			delta := *metric.Delta
			copyM.Delta = &delta
		}
		if metric.Value != nil {
			value := *metric.Value
			copyM.Value = &value
		}
		result = append(result, copyM)
	}
	return result
}
