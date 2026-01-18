package dto

// MetricValueResponse представляет ответ с значением метрики.
// Используется в ответах эндпоинта POST /value.
//
// Поля:
//   - ID: имя метрики
//   - MType: тип метрики
//   - Value: значение для gauge метрики (nil для counter)
//   - Delta: значение для counter метрики (nil для gauge)
type MetricValueResponse struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
