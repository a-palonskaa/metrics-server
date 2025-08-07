package errhandlers_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	errhandlers "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
)

func TestRetriableErrHadler_Success(t *testing.T) {
	count := 0
	res, err := errhandlers.RetriableErrHadler(func() (string, error) {
		count++
		return "ok", nil
	}, func(err error) bool { return true })

	assert.NoError(t, err)
	assert.Equal(t, "ok", res)
	assert.Equal(t, 1, count)
}

func TestRetriableErrHadler_RetryableError(t *testing.T) {
	count := 0
	start := time.Now()
	errTest := errors.New("error")
	_, err := errhandlers.RetriableErrHadler(func() (string, error) {
		count++
		return "", errTest
	}, func(err error) bool { return true })

	duration := time.Since(start)

	assert.Error(t, err)
	assert.Equal(t, 3, count)
	assert.GreaterOrEqual(t, duration, 9*time.Second)
}

func TestRetriableErrHadler_NonRetryableError(t *testing.T) {
	count := 0
	errTest := errors.New("error")
	_, err := errhandlers.RetriableErrHadler(func() (string, error) {
		count++
		return "", errTest
	}, func(err error) bool { return false })

	assert.Error(t, err)
	assert.Equal(t, 1, count)
}

func TestRetriableErrHadlerVoid(t *testing.T) {
	errTest := errors.New("error")
	err := errhandlers.RetriableErrHadlerVoid(func() error {
		return errTest
	}, func(err error) bool { return true })

	assert.Error(t, err)
}

func TestCompareErrAgent(t *testing.T) {
	assert.True(t, errhandlers.CompareErrAgent(errors.New("error")))
}
