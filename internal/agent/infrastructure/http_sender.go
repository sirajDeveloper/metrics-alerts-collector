package infrastructure

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
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

	jsonData, err := json.Marshal(&req)
	if err != nil {
		return err
	}

	compressedData, err := compressBody(jsonData, "application/json")
	if err != nil {
		return err
	}

	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(compressedData).
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

func compressBody(data []byte, contentType string) ([]byte, error) {
	allowedTypes := []string{"application/json", "text/html"}
	allowed := false
	for _, t := range allowedTypes {
		if contentType == t {
			allowed = true
			break
		}
	}

	if !allowed {
		return data, nil
	}

	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)

	_, err := zw.Write(data)
	if err != nil {
		zw.Close()
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
