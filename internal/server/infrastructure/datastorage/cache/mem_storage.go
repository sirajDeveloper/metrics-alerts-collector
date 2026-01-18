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

// MemStorage реализует хранение метрик в памяти с использованием map.
// Потокобезопасна благодаря sync.RWMutex.
// Реализует интерфейс MetricRepository.
type MemStorage struct {
	sync.RWMutex
	cache map[metricKey]*model.Metrics
}

var _ repository.MetricRepository = (*MemStorage)(nil)

// NewMemStorage создает новое хранилище метрик в памяти.
//
// Возвращает новый экземпляр MemStorage с пустым кешем.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		cache: make(map[metricKey]*model.Metrics),
	}
}

// GetMetric возвращает метрику по типу и имени.
// Возвращает копию метрики для защиты от изменения исходных данных.
//
// Параметры:
//   - mType: тип метрики ("gauge" или "counter")
//   - name: имя метрики
//
// Возвращает указатель на копию метрики или nil, если метрика не найдена.
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
		delta := *metric.Delta
		result.Delta = &delta

	}
	if metric.Value != nil {
		value := *metric.Value
		result.Value = &value
	}
	return result
}

// Save сохраняет метрику в кеше.
// Если метрика с таким типом и именем уже существует, она будет перезаписана.
//
// Параметры:
//   - metrics: указатель на метрику для сохранения
func (m *MemStorage) Save(metrics *model.Metrics) {
	m.Lock()
	defer m.Unlock()
	key := metricKey{ID: metrics.ID, MType: metrics.MType}
	m.cache[key] = metrics
}

// GetAll возвращает все метрики из кеша.
// Возвращает копии метрик для защиты от изменения исходных данных.
//
// Возвращает слайс указателей на метрики или nil, если кеш пуст.
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
