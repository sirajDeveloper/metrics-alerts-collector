package event

import (
	"context"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type AuditObserver interface {
	Handle(ctx context.Context, event *model.AuditEvent) error
}
