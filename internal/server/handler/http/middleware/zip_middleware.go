package middleware

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/sirajDeveloper/metrics-alerts-collector/internal/server/infrastructure/compress"
)

func GzipMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get(compress.AcceptEncoding)
		contentEncoding := r.Header.Get(compress.ContentEncoding)

		if strings.Contains(contentEncoding, compress.ContentEncodingGzip) {
			contentType, _, err := mime.ParseMediaType(r.Header.Get(compress.ContentType))
			if err == nil {
				allowContentType := contentType == compress.ContentTypeJSON || contentType == compress.ContentTypeHTML
				if allowContentType {
					zr, err := compress.NewGzipReader(r.Body)
					if err != nil {
						http.Error(w, "invalid gzip data", http.StatusBadRequest)
						return
					}
					defer zr.Close()
					r.Body = io.NopCloser(zr)
				}
			}
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
