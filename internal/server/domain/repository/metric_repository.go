package repository

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

// MetricRepository определяет интерфейс для хранения метрик.
// Реализуется MemStorage и MetricsPostgresRepository.
type MetricRepository interface {
	// Save сохраняет метрику в хранилище.
	// Если метрика с таким типом и именем уже существует, она будет перезаписана.
	Save(metrics *model.Metrics)
	// GetAll возвращает все метрики из хранилища.
	// Возвращает nil, если хранилище пусто.
	GetAll() []*model.Metrics
	// GetMetric возвращает метрику по типу и имени.
	// Возвращает nil, если метрика не найдена.
	GetMetric(mType, name string) *model.Metrics
}

// MetricFileStorage определяет интерфейс для сохранения метрик в файл.
// Используется для персистентного хранения метрик.
type MetricFileStorage interface {
	// SaveAll сохраняет все метрики в файл.
	SaveAll(metrics []*model.Metrics)
	// Save сохраняет одну метрику в файл.
	Save(metric *model.Metrics)
	// LoadAll загружает все метрики из файла.
	// Возвращает ошибку, если файл не найден или поврежден.
	LoadAll() ([]*model.Metrics, error)
}
