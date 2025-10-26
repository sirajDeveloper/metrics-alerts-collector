package dto

type MetricValueRequest struct {
	Name string `json:"id" validate:"required"`
	Type string `json:"type" validate:"required"`
}
