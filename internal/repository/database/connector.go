// Package database provides utilities for establishing and verifying
// a connection to a PostgreSQL database using the pgx driver.
package database

import (
	"context"
	"database/sql"

	"github.com/rs/zerolog/log"

	errhandlers "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
)

// DBConnector wraps a *sql.DB instance and provides methods for interacting
// with the database connection.
type DBConnector struct {
	db *sql.DB
}

// NewConn initializes a new DBConnector with a connection to the given
// PostgreSQL database address.
func NewConn(databaseAddr string) DBConnector {
	db, err := errhandlers.RetriableErrHadler(
		func() (*sql.DB, error) { return sql.Open("pgx", databaseAddr) },
		errhandlers.CompareErrSQL,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize *sql.DB and create a connection pull")
		db = nil
	}
	return DBConnector{
		db: db,
	}
}

func (db DBConnector) Ping(ctx context.Context) error {
	if err := db.db.PingContext(ctx); err != nil {
		log.Error().Err(err).Msg("database ping failed")
		return err
	}
	return nil
}
