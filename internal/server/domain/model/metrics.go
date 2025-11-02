package model

import (
	"errors"
	"strconv"
)

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type MetricType string

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func CreateMetric(id, mType string) *Metrics {
	return &Metrics{
		ID:    id,
		MType: mType,
	}
}

func (m *Metrics) UpdateMetric(newValue any) error {
	switch m.MType {
	case string(Gauge):
		return m.updateGauge(newValue)
	case string(Counter):
		return m.updateCounter(newValue)
	default:
		return errors.New("invalid metric type: " + m.MType)
	}
}

func (m *Metrics) updateGauge(val any) error {
	var floatVal float64
	switch v := val.(type) {
	case float64:
		floatVal = v
	case float32:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	case int:
		floatVal = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return errors.New("updateGauge: invalid float64 value")
		}
		floatVal = parsed
	default:
		return errors.New("updateGauge: unsupported value type")
	}
	m.Value = &floatVal
	return nil
}

func (m *Metrics) updateCounter(val any) error {
	var intVal int64
	switch v := val.(type) {
	case int64:
		intVal = v
	case int:
		intVal = int64(v)
	case float64:
		intVal = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return errors.New("updateCounter: invalid int64 value")
		}
		intVal = parsed
	default:
		return errors.New("updateCounter: unsupported value type")
	}

	total := intVal
	if m.Delta != nil {
		total += *m.Delta
	}
	m.Delta = &total
	return nil
}
