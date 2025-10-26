package dto

type MetricValueRequest struct {
	Name string `json:"name" validate:"required"`
	Type string `json:"type" validate:"required"`
}
