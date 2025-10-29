package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	ContentEncoding     = "Content-Encoding"
	ContentType         = "Content-Type"
	AcceptEncoding      = "Accept-Encoding"
	ContentEncodingGzip = "gzip"
	ContentTypeJSON     = "application/json"
	ContentTypeHTML     = "text/html"
)

type gzipReader struct {
	io.ReadCloser
	reader *gzip.Reader
}

func (gr *gzipReader) Read(p []byte) (int, error) {
	return gr.reader.Read(p)
}

func (gr *gzipReader) Close() error {
	if err := gr.ReadCloser.Close(); err != nil {
		return err
	}
	return gr.reader.Close()
}

func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &gzipReader{
		r,
		reader,
	}, nil
}

type GzipWriter struct {
	http.ResponseWriter
	Writer *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		ResponseWriter: w,
		Writer:         gzip.NewWriter(w),
	}
}

func (gw *GzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		contentType := gw.Header().Get(ContentType)
		if strings.Contains(contentType, ContentTypeJSON) || strings.Contains(contentType, ContentTypeHTML) {
			if gw.Header().Get(ContentEncoding) == "" {
				gw.Header().Set(ContentEncoding, ContentEncodingGzip)
				gw.Writer = gzip.NewWriter(gw.ResponseWriter)
			}
		}
	}
	gw.ResponseWriter.WriteHeader(statusCode)
}

func (gw *GzipWriter) Write(b []byte) (int, error) {
	if gw.Header().Get(ContentEncoding) != "" && gw.Writer != nil {
		return gw.Writer.Write(b)
	}
	return gw.ResponseWriter.Write(b)
}

func (gw *GzipWriter) Close() error {
	if gw.Writer != nil {
		return gw.Writer.Close()
	}
	return nil
}
