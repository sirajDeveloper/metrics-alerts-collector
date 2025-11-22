package infrastructure

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"

	"github.com/go-resty/resty/v2"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"
)

type HTTPSender struct {
	serverURL  string
	client     *resty.Client
	secretKey  string
	retryCount int
}

func NewHTTPSender(url, secretKey string, retryCount int) *HTTPSender {
	return &HTTPSender{
		serverURL:  url,
		client:     resty.New(),
		secretKey:  secretKey,
		retryCount: retryCount,
	}
}

func (s *HTTPSender) executeWithRetry(operation func() error) error {
	for attempt := 1; attempt <= s.retryCount; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		if attempt < s.retryCount {
			delay := time.Duration(2*attempt-1) * time.Second
			logger.Log.Warn("operation failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", s.retryCount),
				zap.Duration("delay", delay),
				zap.Error(err))
			time.Sleep(delay)
			continue
		}

		logger.Log.Error("operation failed after all retry attempts",
			zap.Int("attempts", s.retryCount),
			zap.Error(err))
		return err
	}

	return fmt.Errorf("operation failed")
}

func (s *HTTPSender) sendRequest(url string, reqBody interface{}) error {
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	compressedData, err := compressBody(jsonData, "application/json")
	if err != nil {
		return err
	}

	secret := []byte(s.secretKey)
	mac := hmac.New(sha256.New, secret)
	mac.Write(jsonData)
	hash := mac.Sum(nil)
	hashHex := hex.EncodeToString(hash)

	return s.executeWithRetry(func() error {
		logger.Log.Info("Request to", zap.String("url", url))

		resp, err := s.client.R().
			SetHeader("Content-Type", "application/json").
			SetHeader("Content-Encoding", "gzip").
			SetHeader("HashSHA256", hashHex).
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
	})
}

var _ usecase.MetricSender = (*HTTPSender)(nil)

func (s *HTTPSender) Send(metric *domain.Metric) {
	req, err := s.prepareMetricRequest(metric)
	if err != nil {
		logger.Log.Error("metric preparation failed", zap.String("metric", metric.Name), zap.Error(err))
		return
	}

	url := s.serverURL + "/update"
	if err := s.sendRequest(url, req); err != nil {
		logger.Log.Error("batch send failed", zap.String("url", url), zap.Error(err))
	}
}

var _ usecase.MetricBatchSender = (*HTTPSender)(nil)

func (s *HTTPSender) SendBatch(metrics []domain.Metric) {
	reqs := make([]MetricUpdateRequest, 0, len(metrics))

	for i := range metrics {
		metric := metrics[i]
		req, err := s.prepareMetricRequest(&metric)
		if err != nil {
			logger.Log.Error("metric preparation failed", zap.String("metric", metric.Name), zap.Error(err))
			return
		}
		reqs = append(reqs, *req)
	}

	url := s.serverURL + "/updates"
	if err := s.sendRequest(url, reqs); err != nil {
		logger.Log.Error("batch send failed", zap.String("url", url), zap.Error(err))
	}
}

func (s *HTTPSender) prepareMetricRequest(metric *domain.Metric) (*MetricUpdateRequest, error) {
	var req MetricUpdateRequest

	switch metric.Type {
	case domain.Counter:
		valInt, ok := metric.Value.(int64)
		if !ok {
			return nil, fmt.Errorf("invalid counter value type")
		}
		req = MetricUpdateRequest{
			Name:  metric.Name,
			Type:  string(metric.Type),
			Delta: &valInt,
		}
	case domain.Gauge:
		valFloat, ok := metric.Value.(float64)
		if !ok {
			return nil, fmt.Errorf("invalid gauge value type")
		}
		req = MetricUpdateRequest{
			Name:  metric.Name,
			Type:  string(metric.Type),
			Value: &valFloat,
		}
	default:
		return nil, fmt.Errorf("unknown metric type")
	}
	return &req, nil
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
