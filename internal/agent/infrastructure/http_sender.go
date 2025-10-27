package infrastructure

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"
)

type HTTPSender struct {
	serverURL string
	client    *resty.Client
}

func NewHTTPSender(url string) *HTTPSender {
	return &HTTPSender{
		serverURL: url,
		client:    resty.New(),
	}
}

var _ usecase.MetricSender = (*HTTPSender)(nil)

func (s *HTTPSender) Send(metric domain.Metric) error {
	var req MetricUpdateRequest

	switch metric.Type {
	case domain.Counter:
		valInt, ok := metric.Value.(int64)
		if !ok {
			return fmt.Errorf("invalid counter value type")
		}
		req = MetricUpdateRequest{
			Name:  metric.Name,
			Type:  string(metric.Type),
			Delta: &valInt,
		}
	case domain.Gauge:
		valFloat, ok := metric.Value.(float64)
		if !ok {
			return fmt.Errorf("invalid gauge value type")
		}
		req = MetricUpdateRequest{
			Name:  metric.Name,
			Type:  string(metric.Type),
			Value: &valFloat,
		}
	default:
		return fmt.Errorf("unknown metric type")
	}

	url := s.serverURL + "/update"
	logger.Log.Info("Request to", zap.String("url", url), zap.Any("body", req))

	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(&req).
		Post(url)
	if err != nil {
		return err
	}
	logger.Log.Info("Response", zap.Int("http code", resp.StatusCode()))

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return nil
}

type MetricUpdateRequest struct {
	Name  string   `json:"id" validate:"required"`
	Type  string   `json:"type" validate:"required"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
