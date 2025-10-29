package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

type GzipWriter struct {
	http.ResponseWriter
	zw          *gzip.Writer
	wroteHeader bool
	statusCode  int
}

func NewGzipWriter(w http.ResponseWriter) *GzipWriter {
	return &GzipWriter{
		w,
		gzip.NewWriter(w),
		false,
		0,
	}
}

func (c *GzipWriter) Header() http.Header {
	return c.ResponseWriter.Header()
}

func (c *GzipWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	if c.statusCode < 300 {
		return c.zw.Write(p)
	}
	return c.ResponseWriter.Write(p)
}

func (c *GzipWriter) WriteHeader(statusCode int) {
	if !c.wroteHeader {
		c.statusCode = statusCode
		if statusCode < 300 {
			c.Header().Set("Content-Encoding", "gzip")
		}
		c.wroteHeader = true
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *GzipWriter) Close() error {
	if c.statusCode < 300 {
		return c.zw.Close()
	}
	return nil
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
