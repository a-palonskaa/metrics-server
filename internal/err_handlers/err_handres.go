package errhandlers

import (
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/rs/zerolog/log"
)

// args ...interface{}
// ХУЙНЯ
func RetriableErrHadlerRV[T any](f func() (T, error), compare func(error) bool) (T, error) {
	backoffScedule := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	var (
		args T
		err  error
	)

	for _, backoff := range backoffScedule {
		args, err = f()
		if errors.Is(err, nil) {
			return args, nil
		}
		log.Error().Err(err)

		if !compare(err) {
			return args, err
		}
		time.Sleep(backoff)
	}
	return args, err
}

func RetriableErrHadler(f func() error, compare func(error) bool) error {
	backoffScedule := []time.Duration{
		1 * time.Second,
		3 * time.Second,
		5 * time.Second,
	}

	var (
		err error
	)

	for _, backoff := range backoffScedule {
		err = f()
		if errors.Is(err, nil) {
			return nil
		}
		log.Error().Err(err)

		retry := compare(err)
		if !retry {
			return err
		}
		time.Sleep(backoff)
	}
	return err
}

func CompareErrAgent(err error) bool {
	return true
}

func CompareErrSQL(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return (pgErr.Code == pgerrcode.ConnectionException) ||
			(pgErr.Code == pgerrcode.ConnectionDoesNotExist) ||
			(pgErr.Code == pgerrcode.ConnectionFailure) ||
			(pgErr.Code == pgerrcode.SQLClientUnableToEstablishSQLConnection) ||
			(pgErr.Code == pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection) ||
			(pgErr.Code == pgerrcode.TransactionResolutionUnknown) ||
			(pgErr.Code == pgerrcode.ProtocolViolation)
	}
	return false
}
