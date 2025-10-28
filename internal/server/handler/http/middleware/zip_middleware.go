package middleware

import (
	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/compress"
	"net/http"
	"strings"
)

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		contentType := r.Header.Get("Content-Type")
		allowContentType := strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")

		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if allowContentType && supportsGzip {
			cw := compress.NewGzipWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if allowContentType && sendsGzip {
			cr, err := compress.NewGzipReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	})
}
