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

func (m *Metrics) UpdateMetric(newValue string) error {
	switch m.MType {
	case string(Gauge):
		return m.updateGauge(newValue)
	case string(Counter):
		return m.updateCounter(newValue)
	default:
		return errors.New("invalid metric type: " + m.MType)
	}
}

func (m *Metrics) updateGauge(val string) error {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return errors.New("updateGauge: invalid float64 value")
	}
	m.Value = &v
	return nil
}

func (m *Metrics) updateCounter(val string) error {
	v, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return errors.New("updateCounter: invalid int64 value")
	}
	total := v
	if m.Delta != nil {
		total += *m.Delta
	}
	m.Delta = &total
	return nil
}
