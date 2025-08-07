package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type gzipResponseWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	return w.gz.Write(b)
}

func (w *gzipResponseWriter) Close() error {
	return w.gz.Close()
}

var gzipWriterPool = sync.Pool{
	New: func() any {
		gz, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		if err != nil {
			log.Error().Err(err).Msg("failed to create gzip writer")
		}
		return gz
	},
}

var gzipReaderPool = sync.Pool{
	New: func() any {
		return new(gzip.Reader)
	},
}

func WithCompression(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz := gzipReaderPool.Get().(*gzip.Reader)
			err := gz.Reset(r.Body)
			if err != nil {
				log.Error().Err(err).Msg("failed to reset gzip reader")
				http.Error(w, "", http.StatusBadRequest)
				gzipReaderPool.Put(gz)
				return
			}
			defer func() {
				if err := gz.Close(); err != nil {
					log.Error().Err(err).Msg("error closing gzip reader")
				}
				gzipReaderPool.Put(gz)
			}()
			r.Body = gz
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			fn.ServeHTTP(w, r)
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(w)
		defer func() {
			if err := gz.Close(); err != nil {
				log.Error().Err(err).Msg("error closing gzip writer")
			}
			gzipWriterPool.Put(gz)
		}()
		w.Header().Set("Content-Encoding", "gzip")

		gzw := &gzipResponseWriter{
			ResponseWriter: w,
			gz:             gz,
		}
		fn.ServeHTTP(gzw, r)
	})
}
