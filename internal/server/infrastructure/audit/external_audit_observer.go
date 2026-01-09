package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type externalAuditObserver struct {
	baseURL    string
	httpClient *http.Client
}

func NewExternalAuditObserver(baseURL string) event.AuditObserver {
	return &externalAuditObserver{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (o *externalAuditObserver) Handle(ctx context.Context, auditEvent *model.AuditEvent) error {
	payload, err := json.Marshal(auditEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	url := fmt.Sprintf("%s/api/audit", o.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send audit event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("audit service returned error status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
