package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func WithCompression(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decompress request", http.StatusBadRequest)
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					log.Fatal().Err(err)
				}
			}()
			r.Body = gz
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			log.Error().Err(err).Msg("failed to create gzip writer")
			fn(w, r)
			return
		}
		defer func() {
			if err := gz.Close(); err != nil {
				log.Fatal().Err(err)
			}
		}()

		w.Header().Set("Content-Encoding", "gzip")
		fn(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	}
}
