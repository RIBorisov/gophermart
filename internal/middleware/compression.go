package middleware

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/RIBorisov/gophermart/internal/logger"
)

type compressWriter struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{w, gzip.NewWriter(w)}
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	const gzipThreshold = 300
	if statusCode < gzipThreshold {
		c.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *compressWriter) Close() error {
	return c.zw.Close()
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("failed to create new reader: %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (int, error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	err1 := c.r.Close()
	err2 := c.zr.Close()
	return errors.Join(err1, err2)
}

type BaseMW struct {
	Log *logger.Log
}

func Compression(log *logger.Log) *BaseMW {
	return &BaseMW{
		Log: log,
	}
}

func (g *BaseMW) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		supportsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
		if supportsGzip {
			cw := newCompressWriter(w)
			ow = cw
			defer func() {
				err := cw.Close()
				if err != nil {
					g.Log.Err("failed to close compress writer", err)
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
			}()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				g.Log.Err("failed to read compressed body", err)
				http.Error(w, "check if gzip data is valid", http.StatusBadRequest)
				return
			} else {
				r.Body = cr
				defer func() {
					err = cr.Close()
					if err != nil {
						g.Log.Err("failed to close compress reader", err)
						http.Error(w, "", http.StatusInternalServerError)
						return
					}
				}()
			}
		}
		next.ServeHTTP(ow, r)
	})
}
