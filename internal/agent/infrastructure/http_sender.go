package infrastructure

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/domain"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/agent/usecase"
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
	fmt.Printf("Request url: %v\n", url)

	resp, err := s.client.R().
		SetHeader("Content-Type", "text/plain").
		Post(url)
	if err != nil {
		return err
	}
	fmt.Printf("Response http code: %v\n", resp.StatusCode())

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	/*	req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "text/plain")

		fmt.Printf("Request to: %v metric: %v\n", req.URL.Path, metric)
		resp, err := s.client.Do(req)
		if err != nil {
			return err
		}

		fmt.Printf("Response http code: %v\n", resp.StatusCode)

		defer func() {
			if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
				err = closeErr
			}
		}()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
	*/
	return nil
}
