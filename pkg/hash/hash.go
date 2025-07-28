package hash

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/rs/zerolog/log"
)

func Calculate(key, data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	if _, err := h.Write(data); err != nil {
		log.Error().Err(err).Msg("failed to write HMAC data")
		return nil, err
	}
	return h.Sum(nil), nil
}

func Verify(key, data, expected []byte) bool {
	actual, err := Calculate(key, data)
	if err != nil {
		log.Error().Err(err).Msg("failed to compute HMAC hash")
		return false
	}
	return hmac.Equal(actual, expected)
}
