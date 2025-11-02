package dto

type MetricValueRequest struct {
	ID    string `json:"id" validate:"required"`
	MType string `json:"type" validate:"required"`
}
