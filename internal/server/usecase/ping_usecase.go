package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/rs/zerolog/log"

	errhandlers "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
)

var (
	ErrDBNotInitialized = errors.New("database does not exist")
)

type PingUsecase struct {
	db *sql.DB
}

func NewPingUsecase(databaseAddr string) PingUsecase {
	db, err := errhandlers.RetriableErrHadler(
		func() (*sql.DB, error) { return sql.Open("pgx", databaseAddr) },
		errhandlers.CompareErrSQL,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize *sql.DB and create a connection pull")
		db = nil
	}
	return PingUsecase{
		db: db,
	}
}

func (pu PingUsecase) PingContext(ctx context.Context) error {
	if pu.db == nil {
		return ErrDBNotInitialized
	}
	return pu.db.PingContext(ctx)
}
