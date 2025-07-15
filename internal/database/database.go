package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"runtime"

	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
)

type MyDB struct {
	DB *sql.DB
}

func CreateTables(db *sql.DB) error {
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
}

//----------------------mem-storage interface----------------------

func (db MyDB) IsGaugeAllowed(name string) bool {
	rows, err := db.DB.Query("SELECT * FROM GaugeMetrics WHERE ID = $1", name)
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

func (db MyDB) IsCounterAllowed(name string) bool {
	rows, err := db.DB.Query("SELECT * FROM CounterMetrics WHERE ID = $1", name)
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

func (db MyDB) IsNameAllowed(mType, name string) bool {
	switch mType {
	case metrics.GaugeName:
		return db.IsGaugeAllowed(name)
	case metrics.CounterName:
		return db.IsCounterAllowed(name)
	default:
		log.Error().Msgf("unallowed type %s", mType)
		return false
	}
}

func (db MyDB) AddGauge(name string, val metrics.Gauge) {
	_, err := db.DB.Exec(`
		INSERT INTO GaugeMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value
		`, name, float64(val),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to add gauge metric")
	}
}

func (db MyDB) AddCounter(name string, val metrics.Counter) {
	_, err := db.DB.Exec(`
		INSERT INTO CounterMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = CounterMetrics.Value + EXCLUDED.Value
		`, name, int64(val),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment counter metric")
	}
}

func (db MyDB) GetGaugeValue(name string) (metrics.Gauge, bool) {
	row := db.DB.QueryRow("SELECT Value FROM GaugeMetrics WHERE ID = $1", name)
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

func (db MyDB) GetCounterValue(name string) (metrics.Counter, bool) {
	row := db.DB.QueryRow("SELECT Value FROM CounterMetrics WHERE ID = $1", name)
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

func AddGaugeTx(tx *sql.Tx, name string, val metrics.Gauge) {
	_, err := tx.Exec(`
		INSERT INTO GaugeMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value
		`, name, float64(val),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to add gauge metric")
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err).Msg("failed to roll back")
		}
	}
}

func AddCounterTx(tx *sql.Tx, name string, val metrics.Counter) {
	_, err := tx.Exec(`
		INSERT INTO CounterMetrics (ID, Value)
        VALUES ($1, $2)
        ON CONFLICT (ID)
        DO UPDATE SET Value = CounterMetrics.Value + EXCLUDED.Value
		`, name, int64(val),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment counter metric")
		if err := tx.Rollback(); err != nil {
			log.Error().Err(err).Msg("failed to roll back")
		}
	}
}

func (db MyDB) Update(memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	tx, err := db.DB.Begin()
	if err != nil {
		log.Error().Err(err)
		return
	}

	AddGaugeTx(tx, "Alloc", metrics.Gauge(memStats.Alloc))
	AddGaugeTx(tx, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	AddGaugeTx(tx, "Frees", metrics.Gauge(memStats.Frees))
	AddGaugeTx(tx, "GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	AddGaugeTx(tx, "GCSys", metrics.Gauge(memStats.GCSys))
	AddGaugeTx(tx, "HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	AddGaugeTx(tx, "HeapIdle", metrics.Gauge(memStats.HeapIdle))
	AddGaugeTx(tx, "HeapInuse", metrics.Gauge(memStats.HeapInuse))
	AddGaugeTx(tx, "HeapObjects", metrics.Gauge(memStats.HeapObjects))
	AddGaugeTx(tx, "HeapReleased", metrics.Gauge(memStats.HeapReleased))
	AddGaugeTx(tx, "LastGC", metrics.Gauge(memStats.LastGC))
	AddGaugeTx(tx, "Lookups", metrics.Gauge(memStats.Lookups))
	AddGaugeTx(tx, "MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	AddGaugeTx(tx, "MCacheSys", metrics.Gauge(memStats.MCacheSys))
	AddGaugeTx(tx, "MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	AddGaugeTx(tx, "MSpanSys", metrics.Gauge(memStats.MSpanSys))
	AddGaugeTx(tx, "Mallocs", metrics.Gauge(memStats.Mallocs))
	AddGaugeTx(tx, "NextGC", metrics.Gauge(memStats.NextGC))
	AddGaugeTx(tx, "NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	AddGaugeTx(tx, "NumGC", metrics.Gauge(memStats.NumGC))
	AddGaugeTx(tx, "OtherSys", metrics.Gauge(memStats.OtherSys))
	AddGaugeTx(tx, "PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	AddGaugeTx(tx, "StackInuse", metrics.Gauge(memStats.StackInuse))
	AddGaugeTx(tx, "StackSys", metrics.Gauge(memStats.StackSys))
	AddGaugeTx(tx, "Sys", metrics.Gauge(memStats.Sys))
	AddGaugeTx(tx, "TotalAlloc", metrics.Gauge(memStats.TotalAlloc))
	AddGaugeTx(tx, "HeapSys", metrics.Gauge(memStats.HeapSys))
	AddGaugeTx(tx, "RandomValue", metrics.Gauge(rand.Float64()))
	AddCounterTx(tx, "PollCount", metrics.Counter(1))

	if err := tx.Commit(); err != nil {
		log.Error().Err(err)
	}
}

func (db MyDB) Iterate(f func(string, string, fmt.Stringer)) {
	rowsGauge, err := db.DB.Query("SELECT ID, Value FROM GaugeMetrics")
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

	rowsCounter, err := db.DB.Query("SELECT ID, Value FROM CounterMetrics")
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

//----------------------sex----------------------
