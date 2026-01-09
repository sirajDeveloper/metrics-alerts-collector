package audit

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/event"
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/domain/model"
)

type auditEventPublisher struct {
	observers []event.AuditObserver
	mu        sync.RWMutex
}

func NewAuditEventPublisher() event.AuditEventPublisher {
	return &auditEventPublisher{
		observers: make([]event.AuditObserver, 0),
	}
}

func (p *auditEventPublisher) Subscribe(observer event.AuditObserver) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers = append(p.observers, observer)
}

func (p *auditEventPublisher) Unsubscribe(observer event.AuditObserver) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, obs := range p.observers {
		if obs == observer {
			p.observers = append(p.observers[:i], p.observers[i+1:]...)
			break
		}
	}
}

func (p *auditEventPublisher) Publish(ctx context.Context, auditEvent *model.AuditEvent) error {
	p.mu.RLock()
	observers := make([]event.AuditObserver, len(p.observers))
	copy(observers, p.observers)
	p.mu.RUnlock()

	for _, observer := range observers {
		if err := observer.Handle(ctx, auditEvent); err != nil {
			logger.Log.Error("audit observer failed",
				zap.Error(err),
				zap.String("ip_address", auditEvent.IPAddress),
			)
		}
	}

	return nil
}
