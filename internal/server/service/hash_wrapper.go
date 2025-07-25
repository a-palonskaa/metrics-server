package server

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
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

				h := hmac.New(sha256.New, []byte(key))
				h.Write(body)
				dst := h.Sum(nil)
				hashExpect := hex.EncodeToString(dst)

				if hashExpect != hashStr {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				myWriter := &hashResponseWriter{ResponseWriter: w, bufer: make([]byte, 0)}
				fn.ServeHTTP(myWriter, r)

				h = hmac.New(sha256.New, []byte(key))
				h.Write(myWriter.bufer)
				hash := hex.EncodeToString(h.Sum(nil))

				w.Header().Set("HashSHA256", hash)
				return
			}
			fn.ServeHTTP(w, r)
		})
	}
}
