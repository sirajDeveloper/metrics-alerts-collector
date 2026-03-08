package infrastructure

import (
	"fmt"
	"time"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"
)

func ExecuteWithRetry(retryCount int, operation func() error) error {
	for attempt := 1; attempt <= retryCount; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		if attempt < retryCount {
			delay := time.Duration(2*attempt-1) * time.Second
			logger.Log.Warn("operation failed, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_attempts", retryCount),
				zap.Duration("delay", delay),
				zap.Error(err))
			time.Sleep(delay)
			continue
		}

		logger.Log.Error("operation failed after all retry attempts",
			zap.Int("attempts", retryCount),
			zap.Error(err))
		return err
	}

	return fmt.Errorf("operation failed")
}
