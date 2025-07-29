package hash_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	hash "github.com/a-palonskaa/metrics-server/pkg/hash"
)

func TestCalculate(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid-input",
			key:     []byte("secret key"),
			data:    []byte("test data"),
			wantErr: false,
		},
		{
			name:    "empty-key",
			key:     []byte{},
			data:    []byte("test data"),
			wantErr: false,
		},
		{
			name:    "empty-data",
			key:     []byte("secret key"),
			data:    []byte{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := hash.Calculate(tt.key, tt.data)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	key := []byte("secret key")
	data := []byte("test data")

	validHash, err := hash.Calculate(key, data)
	assert.NoError(t, err)

	tests := []struct {
		name     string
		key      []byte
		data     []byte
		expected []byte
		want     bool
	}{
		{
			name:     "valid-verification",
			key:      key,
			data:     data,
			expected: validHash,
			want:     true,
		},
		{
			name:     "invalid-key",
			key:      []byte("wrong key"),
			data:     data,
			expected: validHash,
			want:     false,
		},
		{
			name:     "invalid-data",
			key:      key,
			data:     []byte("wrong data"),
			expected: validHash,
			want:     false,
		},
		{
			name:     "empty-expected-hash",
			key:      key,
			data:     data,
			expected: []byte{},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hash.Verify(tt.key, tt.data, tt.expected)
			assert.Equal(t, tt.want, result)
		})
	}
}
