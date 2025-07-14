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
			Value DOUBLE PRECISION
		);

		CREATE TABLE CounterMetrics (
			ID varchar(64) PRIMARY KEY,
			Value BIGINT
		);`)
	return err
}

func (db MyDB) IsGaugeAllowed(name string) bool {
	rows, err := db.DB.Query("SELECT * FROM GaugeMetrics WHERE ID = $1", name)
	if err != nil {
		return false
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	if err := rows.Err(); err != nil {
		return false
	}
	return true
}

func (db MyDB) IsCounterAllowed(name string) bool {
	rows, err := db.DB.Query("SELECT * FROM CounterMetrics WHERE ID = $1", name)
	if err != nil {
		return false
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Error().Err(err)
		}
	}()

	if err := rows.Err(); err != nil {
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
		return false
	}
}

func (db MyDB) AddGauge(name string, val metrics.Gauge) {
	log.Info().Msg("AddingGauge")
	_, err := db.DB.Exec(
		`INSERT INTO GaugeMetrics (ID, Value)
         VALUES ($1, $2)
         ON CONFLICT (ID)
         DO UPDATE SET Value = EXCLUDED.Value`, //SEX
		name,
		float64(val),
	)
	log.Info().Err(err).Msg("Added")
	if err != nil {
		log.Error().Err(err)
		return
	}
}

func (db MyDB) AddCounter(name string, val metrics.Counter) {
	_, err := db.DB.Exec(
		`INSERT INTO CounterMetrics (ID, Value)
         VALUES ($1, $2)
         ON CONFLICT (ID)
         DO UPDATE SET Value = CounterMetrics.Value + EXCLUDED.Value`,
		name,
		int64(val),
	)
	if err != nil {
		log.Error().Err(err).Msg("failed to increment counter metric")
	}
}

func (db MyDB) GetGaugeValue(name string) (metrics.Gauge, bool) {
	row := db.DB.QueryRow("SELECT Value FROM GaugeMetrics WHERE ID = $1", name)
	if row == nil {
		return metrics.Gauge(0), false
	}

	if err := row.Err(); err != nil {
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
		return metrics.Counter(0), false
	}

	if err := row.Err(); err != nil {
		return metrics.Counter(0), false
	}

	valueCounter := int64(0)
	if err := row.Scan(&valueCounter); err != nil {
		log.Error().Err(err)
		return metrics.Counter(0), false
	}
	return metrics.Counter(valueCounter), true
}

func (db MyDB) Update(memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	db.AddGauge("Alloc", metrics.Gauge(memStats.Alloc))
	db.AddGauge("BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	db.AddGauge("Frees", metrics.Gauge(memStats.Frees))
	db.AddGauge("GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	db.AddGauge("GCSys", metrics.Gauge(memStats.GCSys))
	db.AddGauge("HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	db.AddGauge("HeapIdle", metrics.Gauge(memStats.HeapIdle))
	db.AddGauge("HeapInuse", metrics.Gauge(memStats.HeapInuse))
	db.AddGauge("HeapObjects", metrics.Gauge(memStats.HeapObjects))
	db.AddGauge("HeapReleased", metrics.Gauge(memStats.HeapReleased))
	db.AddGauge("LastGC", metrics.Gauge(memStats.LastGC))
	db.AddGauge("Lookups", metrics.Gauge(memStats.Lookups))
	db.AddGauge("MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	db.AddGauge("MCacheSys", metrics.Gauge(memStats.MCacheSys))
	db.AddGauge("MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	db.AddGauge("MSpanSys", metrics.Gauge(memStats.MSpanSys))
	db.AddGauge("Mallocs", metrics.Gauge(memStats.Mallocs))
	db.AddGauge("NextGC", metrics.Gauge(memStats.NextGC))
	db.AddGauge("NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	db.AddGauge("NumGC", metrics.Gauge(memStats.NumGC))
	db.AddGauge("OtherSys", metrics.Gauge(memStats.OtherSys))
	db.AddGauge("PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	db.AddGauge("StackInuse", metrics.Gauge(memStats.StackInuse))
	db.AddGauge("StackSys", metrics.Gauge(memStats.StackSys))
	db.AddGauge("Sys", metrics.Gauge(memStats.Sys))
	db.AddGauge("TotalAlloc", metrics.Gauge(memStats.TotalAlloc))
	db.AddGauge("HeapSys", metrics.Gauge(memStats.HeapSys))
	db.AddGauge("RandomValue", metrics.Gauge(rand.Float64()))
	db.AddCounter("PollCount", metrics.Counter(1))
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

	valueCounter := metrics.Counter(0)
	for rowsCounter.Next() {
		if err := rowsCounter.Scan(&name, &valueCounter); err != nil {
			log.Error().Err(err)
			return
		}
		f(name, metrics.CounterName, valueCounter)
	}
}
