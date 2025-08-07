// Package hash provides utilities for generating and verifying HMAC-SHA256 hashes.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"

	"github.com/rs/zerolog/log"
)

// Calculate computes the HMAC-SHA256 hash of the provided data using the given key.
//
// It returns the resulting hash as a byte slice, or an error if hashing fails.
func Calculate(key, data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	if _, err := h.Write(data); err != nil {
		log.Error().Err(err).Msg("failed to write HMAC data")
		return nil, err
	}
	return h.Sum(nil), nil
}

// Verify checks whether the provided HMAC `expected` matches the calculated HMAC
// of `data` using the given `key`.
//
// It returns true if the hashes are equal, false otherwise.
func Verify(key, data, expected []byte) bool {
	actual, err := Calculate(key, data)
	if err != nil {
		log.Error().Err(err).Msg("failed to compute HMAC hash")
		return false
	}
	return hmac.Equal(actual, expected)
}
