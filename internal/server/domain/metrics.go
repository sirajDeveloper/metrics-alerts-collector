package domain

import (
	"errors"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

func CreateMetric(mType string) *Metrics {
	return &Metrics{
		MType: mType,
	}
}

func (m *Metrics) UpdateMetric(newValue interface{}) error {
	switch m.MType {
	case Gauge:
		return m.updateGauge(newValue)
	case Counter:
		return m.updateCounter(newValue)
	default:
		return errors.New("invalid metric type: " + m.MType)
	}
}

func (m *Metrics) updateGauge(val interface{}) error {
	v, ok := val.(float64)
	if !ok {
		return errors.New("updateGauge: value is not float64")
	}
	m.Value = &v
	return nil
}

func (m *Metrics) updateCounter(val interface{}) error {
	v, ok := val.(int64)
	if !ok {
		return errors.New("updateCounter: value is not int64")
	}
	total := v
	if m.Delta != nil {
		total += *m.Delta
	}
	m.Delta = &total
	return nil
}
