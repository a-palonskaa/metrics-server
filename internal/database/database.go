package database

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"

	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	errhandlers "github.com/a-palonskaa/metrics-server/internal/err_handlers"
	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
)

type MyDB struct {
	DB *sql.DB
}

func CreateTables(db *sql.DB) error {
	return errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := db.Exec(`
		DROP TABLE IF EXISTS GaugeMetrics;
		DROP TABLE IF EXISTS CounterMetrics;

		CREATE TABLE GaugeMetrics (
			ID varchar(64) PRIMARY KEY,
			Value DOUBLE PRECISION NOT NULL
		);

		CREATE TABLE CounterMetrics (
			ID varchar(64) PRIMARY KEY,
			Value BIGINT NOT NULL
		);`)
		return err
	}, errhandlers.CompareErrSQL)
}

//----------------------mem-storage interface----------------------

func (db MyDB) IsGaugeAllowed(ctx context.Context, name string) bool {
	args, err := errhandlers.RetriableErrHadler(func() ([]interface{}, error) {
		rows, err := db.DB.QueryContext(ctx, "SELECT * FROM GaugeMetrics WHERE ID = $1", name)
		if err != nil {
			return []interface{}{nil}, err
		}
		if err = rows.Err(); err != nil {
			log.Error().Err(err)
		}
		return []interface{}{rows}, err
	}, errhandlers.CompareErrSQL)
	rows, _ := args[0].(*sql.Rows)
	if err != nil {
		log.Error().Err(err)
		return false
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	if err := rows.Err(); err != nil {
		log.Error().Err(err)
		return false
	}
	return true
}

func (db MyDB) IsCounterAllowed(ctx context.Context, name string) bool {
	rows, err := errhandlers.RetriableErrHadler(func() (*sql.Rows, error) {
		return db.DB.QueryContext(ctx, "SELECT * FROM CounterMetrics WHERE ID = $1", name)
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
		return false
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	if err := rows.Err(); err != nil {
		log.Error().Err(err)
		return false
	}
	return true
}

func (db MyDB) IsNameAllowed(ctx context.Context, mType, name string) bool {
	switch mType {
	case metrics.GaugeName:
		return db.IsGaugeAllowed(ctx, name)
	case metrics.CounterName:
		return db.IsCounterAllowed(ctx, name)
	default:
		log.Error().Msgf("unallowed type %s", mType)
		return false
	}
}

func (db MyDB) AddGauge(ctx context.Context, name string, val metrics.Gauge) {
	err := errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := db.DB.ExecContext(ctx, `
		INSERT INTO GaugeMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value
		`, name, float64(val),
		)
		return err
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err).Msg("failed to add gauge metric")
	}
}

func (db MyDB) AddCounter(ctx context.Context, name string, val metrics.Counter) {
	err := errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := db.DB.ExecContext(ctx, `
		INSERT INTO CounterMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = CounterMetrics.Value + EXCLUDED.Value
		`, name, int64(val),
		)
		return err
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment counter metric")
	}
}

func (db MyDB) GetGaugeValue(ctx context.Context, name string) (metrics.Gauge, bool) {
	row := db.DB.QueryRowContext(ctx, "SELECT Value FROM GaugeMetrics WHERE ID = $1", name)
	if row == nil {
		log.Info().Msgf("no gauge val with name %s found", name)
		return metrics.Gauge(0), false
	}

	if err := row.Err(); err != nil {
		log.Error().Err(err)
		return metrics.Gauge(0), false
	}

	valueGauge := float64(0)
	if err := row.Scan(&valueGauge); err != nil {
		log.Error().Err(err)
		return metrics.Gauge(0), false
	}
	return metrics.Gauge(valueGauge), true
}

func (db MyDB) GetCounterValue(ctx context.Context, name string) (metrics.Counter, bool) {
	row := db.DB.QueryRowContext(ctx, "SELECT Value FROM CounterMetrics WHERE ID = $1", name)
	if row == nil {
		log.Info().Msgf("no counter val with name %s found", name)
		return metrics.Counter(0), false
	}

	if err := row.Err(); err != nil {
		log.Error().Err(err)
		return metrics.Counter(0), false
	}

	valueCounter := int64(0)
	if err := row.Scan(&valueCounter); err != nil {
		log.Error().Err(err)
		return metrics.Counter(0), false
	}
	return metrics.Counter(valueCounter), true
}

func AddCounterTx(ctx context.Context, tx *sql.Tx, name string, val metrics.Counter) {
	err := errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO CounterMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = CounterMetrics.Value + EXCLUDED.Value
		`, name, int64(val),
		)
		return err
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment counter metric")
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err).Msg("failed to roll back")
		}
	}
}

func AddGaugeTx(ctx context.Context, tx *sql.Tx, name string, val metrics.Gauge) {
	err := errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := tx.ExecContext(ctx, `
		INSERT INTO GaugeMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value
		`, name, float64(val),
		)
		return err
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err).Msg("failed to add gauge metric")
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err).Msg("failed to roll back")
		}
	}
}

func ExecQuery(ctx context.Context, stmt *sql.Stmt, args ...interface{}) {
	err := errhandlers.RetriableErrHadlerVoid(func() error {
		_, err := stmt.ExecContext(ctx, args...)
		return err
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
	}
}

func (db MyDB) Update(ctx context.Context, memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	tx, err := errhandlers.RetriableErrHadler(func() (*sql.Tx, error) {
		return db.DB.Begin()
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err)
		}
		return
	}

	stmt, err := errhandlers.RetriableErrHadler(func() (*sql.Stmt, error) {
		return tx.PrepareContext(ctx, `INSERT INTO GaugeMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value`)
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	ExecQuery(ctx, stmt, "Alloc", metrics.Gauge(memStats.Alloc))
	ExecQuery(ctx, stmt, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	ExecQuery(ctx, stmt, "Frees", metrics.Gauge(memStats.Frees))
	ExecQuery(ctx, stmt, "GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	ExecQuery(ctx, stmt, "GCSys", metrics.Gauge(memStats.GCSys))
	ExecQuery(ctx, stmt, "HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	ExecQuery(ctx, stmt, "HeapIdle", metrics.Gauge(memStats.HeapIdle))
	ExecQuery(ctx, stmt, "HeapInuse", metrics.Gauge(memStats.HeapInuse))
	ExecQuery(ctx, stmt, "HeapObjects", metrics.Gauge(memStats.HeapObjects))
	ExecQuery(ctx, stmt, "HeapReleased", metrics.Gauge(memStats.HeapReleased))
	ExecQuery(ctx, stmt, "LastGC", metrics.Gauge(memStats.LastGC))
	ExecQuery(ctx, stmt, "Lookups", metrics.Gauge(memStats.Lookups))
	ExecQuery(ctx, stmt, "MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	ExecQuery(ctx, stmt, "MCacheSys", metrics.Gauge(memStats.MCacheSys))
	ExecQuery(ctx, stmt, "MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	ExecQuery(ctx, stmt, "MSpanSys", metrics.Gauge(memStats.MSpanSys))
	ExecQuery(ctx, stmt, "Mallocs", metrics.Gauge(memStats.Mallocs))
	ExecQuery(ctx, stmt, "NextGC", metrics.Gauge(memStats.NextGC))
	ExecQuery(ctx, stmt, "NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	ExecQuery(ctx, stmt, "NumGC", metrics.Gauge(memStats.NumGC))
	ExecQuery(ctx, stmt, "OtherSys", metrics.Gauge(memStats.OtherSys))
	ExecQuery(ctx, stmt, "PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	ExecQuery(ctx, stmt, "StackInuse", metrics.Gauge(memStats.StackInuse))
	ExecQuery(ctx, stmt, "StackSys", metrics.Gauge(memStats.StackSys))
	ExecQuery(ctx, stmt, "Sys", metrics.Gauge(memStats.Sys))
	ExecQuery(ctx, stmt, "TotalAlloc", metrics.Gauge(memStats.TotalAlloc))
	ExecQuery(ctx, stmt, "HeapSys", metrics.Gauge(memStats.HeapSys))
	ExecQuery(ctx, stmt, "RandomValue", metrics.Gauge(rand.Float64()))
	AddCounterTx(ctx, tx, "PollCount", metrics.Counter(1))

	err = errhandlers.RetriableErrHadlerVoid(func() error {
		return tx.Commit()
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
	}
}

func (db MyDB) Iterate(ctx context.Context, f func(string, string, fmt.Stringer)) {
	rowsGauge, err := errhandlers.RetriableErrHadler(func() (*sql.Rows, error) {
		return db.DB.QueryContext(ctx, "SELECT ID, Value FROM GaugeMetrics")
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer func() {
		if err := rowsGauge.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	if err := rowsGauge.Err(); err != nil {
		log.Error().Err(err)
		return
	}

	name := ""
	valueGauge := metrics.Gauge(0)
	for rowsGauge.Next() {
		if err := rowsGauge.Scan(&name, &valueGauge); err != nil {
			log.Error().Err(err)
			return
		}
		f(name, metrics.GaugeName, valueGauge)
	}

	rowsCounter, err := errhandlers.RetriableErrHadler(func() (*sql.Rows, error) {
		return db.DB.QueryContext(ctx, "SELECT ID, Value FROM CounterMetrics")
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer func() {
		if err := rowsCounter.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	if err := rowsCounter.Err(); err != nil {
		log.Error().Err(err)
		return
	}

	valueCounter := metrics.Counter(0)
	for rowsCounter.Next() {
		if err := rowsCounter.Scan(&name, &valueCounter); err != nil {
			log.Error().Err(err)
			return
		}
		f(name, metrics.CounterName, valueCounter)
	}
}

//SEX

func (db MyDB) AddMetricsToStorage(ctx context.Context, mt *metrics.MetricsS) int {
	tx, err := errhandlers.RetriableErrHadler(func() (*sql.Tx, error) {
		return db.DB.Begin()
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
		return http.StatusOK
	}

	for _, metric := range *mt {
		switch metric.MType {
		case "gauge":
			AddGaugeTx(ctx, tx, metric.ID, metrics.Gauge(*metric.Value))
		case "counter":
			AddCounterTx(ctx, tx, metric.ID, metrics.Counter(*metric.Delta))
		default:
			return http.StatusBadRequest
		}
	}

	err = errhandlers.RetriableErrHadlerVoid(func() error {
		return tx.Commit()
	}, errhandlers.CompareErrSQL)
	if err != nil {
		log.Error().Err(err)
	}
	return http.StatusOK
}
