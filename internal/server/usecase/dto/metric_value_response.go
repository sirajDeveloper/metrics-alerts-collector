package dto

type MetricValueResponse struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}
