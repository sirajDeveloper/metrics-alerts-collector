package dto

// MetricUpdateRequest представляет запрос на обновление метрики.
// Используется в эндпоинтах POST /update и POST /updates.
//
// Поля:
//   - ID: уникальное имя метрики (обязательное)
//   - MType: тип метрики - "gauge" или "counter" (обязательное)
//   - Value: значение для gauge метрики (обязательно для gauge)
//   - Delta: инкремент для counter метрики (обязательно для counter)
//
// Примеры:
//
//	Gauge: {ID: "temperature", MType: "gauge", Value: 25.5}
//	Counter: {ID: "requests", MType: "counter", Delta: 10}
type MetricUpdateRequest struct {
	ID    string   `json:"id" validate:"required"`
	MType string   `json:"type" validate:"required"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
