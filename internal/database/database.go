package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
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

func ExecQuery(stmt *sql.Stmt, args ...interface{}) {
	_, err := stmt.Exec(args...)
	if err != nil {
		log.Error().Err(err)
	}
}

func (db MyDB) Update(memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	tx, err := db.DB.Begin()
	if err != nil {
		log.Error().Err(err)
		return
	}

	stmt, err := tx.Prepare(`INSERT INTO GaugeMetrics (ID, Value)
        VALUES (?, ?)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value`)
	if err != nil {
		log.Error().Err(err)
		return
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	ExecQuery(stmt, "Alloc", metrics.Gauge(memStats.Alloc))
	ExecQuery(stmt, "Alloc", metrics.Gauge(memStats.Alloc))
	ExecQuery(stmt, "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	ExecQuery(stmt, "Frees", metrics.Gauge(memStats.Frees))
	ExecQuery(stmt, "GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	ExecQuery(stmt, "GCSys", metrics.Gauge(memStats.GCSys))
	ExecQuery(stmt, "HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	ExecQuery(stmt, "HeapIdle", metrics.Gauge(memStats.HeapIdle))
	ExecQuery(stmt, "HeapInuse", metrics.Gauge(memStats.HeapInuse))
	ExecQuery(stmt, "HeapObjects", metrics.Gauge(memStats.HeapObjects))
	ExecQuery(stmt, "HeapReleased", metrics.Gauge(memStats.HeapReleased))
	ExecQuery(stmt, "LastGC", metrics.Gauge(memStats.LastGC))
	ExecQuery(stmt, "Lookups", metrics.Gauge(memStats.Lookups))
	ExecQuery(stmt, "MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	ExecQuery(stmt, "MCacheSys", metrics.Gauge(memStats.MCacheSys))
	ExecQuery(stmt, "MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	ExecQuery(stmt, "MSpanSys", metrics.Gauge(memStats.MSpanSys))
	ExecQuery(stmt, "Mallocs", metrics.Gauge(memStats.Mallocs))
	ExecQuery(stmt, "NextGC", metrics.Gauge(memStats.NextGC))
	ExecQuery(stmt, "NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	ExecQuery(stmt, "NumGC", metrics.Gauge(memStats.NumGC))
	ExecQuery(stmt, "OtherSys", metrics.Gauge(memStats.OtherSys))
	ExecQuery(stmt, "PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	ExecQuery(stmt, "StackInuse", metrics.Gauge(memStats.StackInuse))
	ExecQuery(stmt, "StackSys", metrics.Gauge(memStats.StackSys))
	ExecQuery(stmt, "Sys", metrics.Gauge(memStats.Sys))
	ExecQuery(stmt, "TotalAlloc", metrics.Gauge(memStats.TotalAlloc))
	ExecQuery(stmt, "HeapSys", metrics.Gauge(memStats.HeapSys))
	ExecQuery(stmt, "RandomValue", metrics.Gauge(rand.Float64()))
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

//----------------------sex----------------------  //SEX

func (db MyDB) AddMetricsToStorage(mt *metrics.MetricsS) int {
	tx, err := db.DB.Begin()
	if err != nil {
		log.Error().Err(err)
		return http.StatusOK
	}

	stmt, err := tx.Prepare(`INSERT INTO GaugeMetrics (ID, Value)
        VALUES (?, ?)
        ON CONFLICT (ID)
        DO UPDATE SET Value = EXCLUDED.Value`)
	if err != nil {
		log.Error().Err(err)
		return http.StatusOK
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	for _, metric := range *mt {
		switch metric.MType {
		case "gauge":
			ExecQuery(stmt, metric.ID, metrics.Gauge(*metric.Value))
		case "counter":
			db.AddCounter(metric.ID, metrics.Counter(*metric.Delta))
		default:
			return http.StatusBadRequest
		}
	}

	if err := tx.Commit(); err != nil {
		log.Error().Err(err)
	}
	return http.StatusOK
}
