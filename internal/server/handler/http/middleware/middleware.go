package middleware

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

func LoggingMiddleware(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: writer}

		next.ServeHTTP(rw, request)

		duration := time.Since(start)

		logger.Log.Info("HTTP Request",
			zap.String("method", request.Method),
			zap.String("uri", request.RequestURI),
			zap.Duration("duration", duration),
		)
	})
	return fn
}
