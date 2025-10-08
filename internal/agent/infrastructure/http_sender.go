package infrastructure

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
)

type HTTPSender struct {
	serverURL string
	client    *http.Client
}

func NewHTTPSender(url string) *HTTPSender {
	return &HTTPSender{
		serverURL: url,
		client:    &http.Client{},
	}
}

func (s *HTTPSender) Send(metric domain.Metric) error {
	typeStr := string(metric.Type)

	var valueStr string
	switch v := metric.Value.(type) {
	case float64:
		valueStr = strconv.FormatFloat(v, 'f', -1, 64)
	case int64:
		valueStr = strconv.FormatInt(v, 10)
	case int:
		valueStr = strconv.Itoa(v)
	}

	url := fmt.Sprintf("%s/update/%s/%s/%s", s.serverURL, typeStr, metric.Name, valueStr)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := s.client.Do(req)
	resp.Body.Close()
	return nil
}
