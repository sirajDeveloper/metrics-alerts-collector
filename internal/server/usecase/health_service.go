package usecase

import (
	"context"
	"errors"
)

type DatabaseHealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	checker DatabaseHealthChecker
}

func NewHealthService(checker DatabaseHealthChecker) *HealthService {
	return &HealthService{checker: checker}
}

func (s *HealthService) Ping(ctx context.Context) error {
	if s.checker == nil {
		return errors.New("database health checker is not configured")
	}
	return s.checker.Ping(ctx)
}
