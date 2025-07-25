package server

import (
	"bytes"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"

	hash "github.com/a-palonskaa/metrics-server/pkg/hash"
)

type hashResponseWriter struct {
	http.ResponseWriter
	bufer []byte
}

func (w *hashResponseWriter) Write(bufer []byte) (int, error) {
	w.bufer = bufer
	return w.ResponseWriter.Write(bufer)
}

func CheckHash(key string) func(fn http.Handler) http.Handler {
	return func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" {
				hashStr := ""
				if hashStr = r.Header.Get("HashSHA256"); hashStr == "" {
					fn.ServeHTTP(w, r)
					return
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Error().Err(err).Msg("error reading body")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(body))

				expectedHash, err := hex.DecodeString(hashStr)
				if err != nil {
					log.Error().Err(err).Msg("failed to decode string")
				}
				if !hash.Verify([]byte(key), body, []byte(expectedHash)) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				myWriter := &hashResponseWriter{ResponseWriter: w, bufer: make([]byte, 0)}
				fn.ServeHTTP(myWriter, r)

				dst, err := hash.Calculate([]byte(key), myWriter.bufer)
				if err != nil {
					log.Error().Err(err).Msg("failed to calculate hash")
				}
				w.Header().Set("HashSHA256", hex.EncodeToString(dst))
				return
			}
			fn.ServeHTTP(w, r)
		})
	}
}
