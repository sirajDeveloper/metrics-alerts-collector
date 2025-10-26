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

	req := MetricUpdateRequest{
		Name:  metric.Name,
		Type:  string(metric.Type),
		Value: metric.Value,
	}

	url := s.serverURL + "/update"
	logger.Log.Info("Request to", zap.String("url", url))

	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(&req).
		Post(url)
	if err != nil {
		return err
	}
	logger.Log.Info("Response http code", zap.Int("code", resp.StatusCode()))

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return nil
}

type MetricUpdateRequest struct {
	Name  string `json:"id" validate:"required"`
	Type  string `json:"type" validate:"required"`
	Value any    `json:"value"`
}
