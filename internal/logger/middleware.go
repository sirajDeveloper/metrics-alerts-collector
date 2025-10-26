package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		next.ServeHTTP(w, r)

		duration := time.Since(start)

		Log.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Duration("duration", duration),
		)
	})
}
