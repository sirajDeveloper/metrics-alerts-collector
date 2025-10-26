package dto

type MetricUpdateRequest struct {
	Name  string `json:"id" validate:"required"`
	Type  string `json:"type" validate:"required"`
	Value any    `json:"value"`
}
