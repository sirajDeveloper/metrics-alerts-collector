package dto

// MetricValueRequest представляет запрос на получение значения метрики.
// Используется в эндпоинте POST /value.
//
// Поля:
//   - ID: имя метрики (обязательное)
//   - MType: тип метрики - "gauge" или "counter" (обязательное)
type MetricValueRequest struct {
	ID    string `json:"id" validate:"required"`
	MType string `json:"type" validate:"required"`
}
