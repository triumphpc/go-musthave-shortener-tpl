package middlewares

import (
	"compress/gzip"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

type CompressorMw struct {
	h http.Handler
	l *zap.Logger
}

func NewCompressor(l *zap.Logger) *CompressorMw {
	return &CompressorMw{l: l}
}

// GzipMiddleware compress and decompress zip data
func (h CompressorMw) GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client send gzip format
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {

			fmt.Println("YYY")
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				h.l.Info("decompress error", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}
			defer func(reader *gzip.Reader) {
				_ = reader.Close()
			}(reader)
			r.Body = reader
		}

		//Check if client support gzip for response
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip.Writer
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			h.l.Info("compress error", zap.Error(err))
			next.ServeHTTP(w, r)
			return
		}

		defer func(gz *gzip.Writer) {
			_ = gz.Close()
		}(gz)
		w.Header().Set("Content-Encoding", "gzip")
		// Prepare data
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// Writer response by gzip
	return w.Writer.Write(b)
}
