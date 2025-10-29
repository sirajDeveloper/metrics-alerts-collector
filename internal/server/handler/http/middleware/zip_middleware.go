package middleware

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/logger"
	"go.uber.org/zap"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/compress"
)

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType, _, err := mime.ParseMediaType(r.Header.Get(compress.ContentType))
		if err != nil {
			logger.Log.Error("mime.ParseMediaType has been finished with error. compress won't complete", zap.Error(err))
			h.ServeHTTP(w, r)
			return
		}
		allowContentType := contentType == compress.ContentTypeJSON || contentType == compress.ContentTypeHTML
		acceptEncoding := r.Header.Get(compress.AcceptEncoding)
		contentEncoding := r.Header.Get(compress.ContentEncoding)
		if !allowContentType {
			h.ServeHTTP(w, r)
			return
		}

		if strings.Contains(contentEncoding, compress.ContentEncodingGzip) {
			zr, err := compress.NewGzipReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip data", http.StatusBadRequest)
				return
			}
			defer zr.Close()
			r.Body = io.NopCloser(zr)
		}

		if !strings.Contains(acceptEncoding, compress.ContentEncodingGzip) {
			h.ServeHTTP(w, r)
			return
		}

		gzw := compress.NewGzipWriter(w)
		defer func() {
			if gzw.Writer != nil {
				gzw.Writer.Close()
			}
		}()

		h.ServeHTTP(gzw, r)
	})
}
