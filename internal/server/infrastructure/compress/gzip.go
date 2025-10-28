package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		w,
		gzip.NewWriter(w),
	}
}

func (c *GzipWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *GzipWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *GzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *GzipWriter) Close() error {
	return c.zw.Close()
}

// GzipReader Reader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type GzipReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*GzipReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &GzipReader{
		r,
		zr,
	}, nil
}

func (c *GzipReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *GzipReader) Close() error {
	if err := c.ReadCloser.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func NewCompressWriter(w http.ResponseWriter) *GzipWriter {
	return NewGzipWriter(w)
}

func NewCompressReader(r io.ReadCloser) (*GzipReader, error) {
	return NewGzipReader(r)
}
