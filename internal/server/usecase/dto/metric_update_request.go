package dto

type MetricUpdateRequest struct {
	Name  string `json:"name" validate:"required"`
	Type  string `json:"type" validate:"required"`
	Value string `json:"value"`
}
