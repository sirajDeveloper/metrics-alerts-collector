package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type fileAuditObserver struct {
	filePath string
	file     *os.File
	mu       sync.Mutex
	encoder  *json.Encoder
}

func NewFileAuditObserver(filePath string) (event.AuditObserver, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &fileAuditObserver{
		filePath: filePath,
		file:     file,
		encoder:  json.NewEncoder(file),
	}, nil
}

func (o *fileAuditObserver) Handle(ctx context.Context, auditEvent *model.AuditEvent) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	return o.encoder.Encode(auditEvent)
}

func (o *fileAuditObserver) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.file != nil {
		return o.file.Close()
	}
	return nil
}
