package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressReader struct {
	r      io.ReadCloser
	reader *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	reader, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r: r, reader: reader}, nil
}

type compressWriter struct {
	w           http.ResponseWriter
	writer      *gzip.Writer
	wroteHeader bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w: w}
}

func (c *compressWriter) Header() http.Header { return c.w.Header() }

func (c *compressWriter) WriteHeader(statusCode int) {
	if c.wroteHeader {
		return
	}
	c.wroteHeader = true
	ct := c.w.Header().Get("Content-Type")
	if strings.HasPrefix(ct, "application/json") || strings.HasPrefix(ct, "text/html") {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
		c.writer = gzip.NewWriter(c.w)
	}
	c.w.WriteHeader(statusCode)
}

func (c *compressWriter) Write(p []byte) (int, error) {
	if !c.wroteHeader {
		c.WriteHeader(http.StatusOK)
	}
	if c.writer != nil {
		return c.writer.Write(p)
	}
	return c.w.Write(p)
}

func (c *compressWriter) Close() error {
	if c.writer != nil {
		return c.writer.Close()
	}
	return nil
}

func (c *compressReader) Read(p []byte) (int, error) { return c.reader.Read(p) }

func (c *compressReader) Close() error {
	if err := c.reader.Close(); err != nil {
		return err
	}
	return c.r.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(ow, r)
	})
}
