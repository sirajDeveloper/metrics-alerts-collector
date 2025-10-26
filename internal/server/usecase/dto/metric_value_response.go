package dto

type MetricValueResponse struct {
	Name  string `json:"id"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}
