package event

import (
	"context"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type AuditEventPublisher interface {
	Subscribe(observer AuditObserver)
	Unsubscribe(observer AuditObserver)
	Publish(ctx context.Context, event *model.AuditEvent) error
}
