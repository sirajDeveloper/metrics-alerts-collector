package dto

// DisplayMetricDTO представляет метрику в формате для отображения.
// Используется для генерации HTML страницы со списком метрик.
//
// Поля:
//   - ID: имя метрики
//   - MType: тип метрики ("gauge" или "counter")
//   - ValueStr: значение метрики в виде строки
type DisplayMetricDTO struct {
	ID       string
	MType    string
	ValueStr string
}
