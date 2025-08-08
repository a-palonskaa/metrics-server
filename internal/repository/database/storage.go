// Package database provides a PostgreSQL-based implementation of a metrics repository.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	errhandlers "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
)

// DBStorage is a PostgreSQL-backed metrics storage
type DBStorage struct {
	mu sync.RWMutex
	DB *sql.DB
}

// New initializes a new DBStorage by connecting to the provided PostgreSQL database
// address and creating the required metrics table.
func New(databaseAddr string) (*DBStorage, bool) {
	db, err := errhandlers.RetriableErrHadler(
		func() (*sql.DB, error) { return sql.Open("pgx", databaseAddr) },
		errhandlers.CompareErrSQL,
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize *sql.DB and create a connection pull")
		return nil, false
	}

	err = errhandlers.RetriableErrHadlerVoid(
		func() error { return CreateTables(db) },
		errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err).Msg("failed to create tables")
		return nil, false
	}

	return &DBStorage{
		DB: db,
	}, true
}

func CreateTables(db *sql.DB) error {
	return errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := db.Exec(`
		DROP TABLE IF EXISTS Metrics;

		CREATE TABLE Metrics (
			ID VARCHAR(64) PRIMARY KEY,
			MType VARCHAR(16) NOT NULL,
			Delta BIGINT,
			Value DOUBLE PRECISION
		);`)
		if err != nil {
			return fmt.Errorf("failed to create table:%w", err)
		}
		return nil
	}, errhandlers.CompareErrSQL)
}

func (db *DBStorage) Update(ctx context.Context, metric metrics.Metrics) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.DB.Begin()
	if err != nil {
		log.Error().Err(err).Msg("failed to load transaction")
		return fmt.Errorf("failed to load transaction:%w", err)
	}

	for _, mt := range []metrics.Metric(metric) {
		err := addMetric(tx, mt)
		if err != nil {
			log.Error().Err(err).Msg("failed to add metric")
			if err = tx.Rollback(); err != nil {
				log.Error().Err(err).Msg("failed to rollback")
			}
			return fmt.Errorf("failed to add metric:%w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err).Msg("failed to commit transaction")
		return fmt.Errorf("failed to commit transaction:%w", err)
	}
	return nil
}

func (db *DBStorage) Get(ctx context.Context, mType, name string) (metrics.Metric, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	row := db.DB.QueryRowContext(ctx, `
		SELECT MType, Delta, Value FROM Metrics WHERE ID = $1
	`, name)

	var metric metrics.Metric
	metric.ID = name
	var delta sql.NullInt64
	var value sql.NullFloat64

	if err := row.Err(); err != nil {
		log.Error().Err(err).Msg("error in sql row")
		return metric, err
	}

	if err := row.Scan(&metric.MType, &delta, &value); err != nil {
		return metrics.Metric{}, err
	}

	if metric.MType != mType {
		return metrics.Metric{}, metrics.ErrIncorrectMetricType
	}

	if metric.MType == metrics.GaugeName && value.Valid {
		metric.Value = value.Float64
	} else if metric.MType == metrics.CounterName && delta.Valid {
		metric.Delta = delta.Int64
	}
	return metric, nil
}

func (db *DBStorage) List(ctx context.Context) []metrics.Metric {
	db.mu.Lock()
	defer db.mu.Unlock()

	var allMetrics []metrics.Metric

	rows, err := db.DB.QueryContext(ctx, `SELECT ID, MType, Delta, Value FROM Metrics`)
	if err != nil {
		log.Error().Err(err).Msg("error fetching metrics")
		return allMetrics
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close rows")
		}
	}()

	if err := rows.Err(); err != nil {
		log.Error().Err(err).Msg("error in sql rows")
		return nil
	}

	var metric metrics.Metric
	var delta sql.NullInt64
	var value sql.NullFloat64

	for rows.Next() {
		if err := rows.Scan(&metric.ID, &metric.MType, &delta, &value); err != nil {
			log.Error().Err(err).Msg("error scanning metrics")
			continue
		}

		if metric.MType == metrics.GaugeName && value.Valid {
			metric.Value = value.Float64
		} else if metric.MType == metrics.CounterName && delta.Valid {
			metric.Delta = delta.Int64
		}
		allMetrics = append(allMetrics, metric)
	}
	return allMetrics
}

func (db *DBStorage) Close() error {
	return db.DB.Close()
}

func addMetric(tx *sql.Tx, m metrics.Metric) error {
	switch m.MType {
	case metrics.GaugeName:
		_, err := tx.Exec(`
			INSERT INTO Metrics (ID, MType, Value)
			VALUES ($1, $2, $3)
			ON CONFLICT (ID)
			DO UPDATE SET Value = EXCLUDED.Value, MType = EXCLUDED.MType
		`, m.ID, m.MType, m.Value)
		if err != nil {
			return fmt.Errorf("failed to insert gauge metric:%w", err)
		}
	case metrics.CounterName:
		_, err := tx.Exec(`
			INSERT INTO Metrics (ID, MType, Delta)
			VALUES ($1, $2, $3)
			ON CONFLICT (ID)
			DO UPDATE SET Delta = Metrics.Delta + EXCLUDED.Delta, MType = EXCLUDED.MType
		`, m.ID, m.MType, m.Delta)
		if err != nil {
			return fmt.Errorf("failed to insert counter metric:%w", err)
		}
	default:
		return metrics.ErrIncorrectMetricType
	}
	return nil
}
