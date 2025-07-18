package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

func CheckHash(key string) func(fn http.Handler) http.Handler {
	return func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key != "" {
				hashStr := ""
				if hashStr = r.Header.Get("HashSHA256"); hashStr == "" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					log.Error().Err(err).Msg("error reading body")
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				h := hmac.New(sha256.New, []byte(key))
				h.Write(body)
				dst := h.Sum(nil)
				hashExpect := hex.EncodeToString(dst)

				if hashExpect != hashStr {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			fn.ServeHTTP(w, r)
		})
	}
}
