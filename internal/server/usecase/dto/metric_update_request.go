package dto

type MetricUpdateRequest struct {
	ID    string   `json:"id" validate:"required"`
	MType string   `json:"type" validate:"required"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
