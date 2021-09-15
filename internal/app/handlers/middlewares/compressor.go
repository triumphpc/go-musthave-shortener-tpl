package middlewares

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// GzipMiddleware compress and decompress zip data
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ZIP")
		// Check if client send gzip format
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				log.Println("decompress error", err)
				next.ServeHTTP(w, r)
				return
			}
			defer reader.Close()
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
			log.Println("compress error", err)
			next.ServeHTTP(w, r)
			return
		}

		defer gz.Close()
		w.Header().Set("Content-Encoding", "gzip")
		// Prepare data
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// Writer response by gzip
	return w.Writer.Write(b)
}
