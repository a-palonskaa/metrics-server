package errhandlers

import (
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/rs/zerolog/log"
)

func RetriableErrHadler[T any](f func() (T, error), compare func(error) bool) (T, error) {
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
		if err == nil {
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

func RetriableErrHadlerVoid(f func() error, compare func(error) bool) error {
	_, err := RetriableErrHadler(func() (struct{}, error) {
		return struct{}{}, f()
	}, compare)
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
