package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
)

func TestDBStorage_Get_Gauge(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close db storage")
		}
	}()

	storage := &database.DBStorage{DB: db}
	ctx := context.Background()

	t.Run("correct-case-gauge#1", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"MType", "Delta", "Value"}).
			AddRow("gauge", nil, float64(3.14))

		mock.ExpectQuery(`SELECT MType, Delta, Value FROM Metrics WHERE ID = \$1`).
			WithArgs("load_avg").
			WillReturnRows(rows)

		metric, err := storage.Get(ctx, "gauge", "load_avg")
		require.NoError(t, err)
		require.Equal(t, "load_avg", metric.ID)
		require.Equal(t, "gauge", metric.MType)
		require.Equal(t, float64(3.14), metric.Value)
	})

	t.Run("incorrect-case#1", func(t *testing.T) {
		mock.ExpectQuery(`SELECT MType, Delta, Value FROM Metrics WHERE ID = \$1`).
			WithArgs("unknown").
			WillReturnError(sql.ErrNoRows)

		_, err := storage.Get(ctx, "gauge", "unknown")
		require.Error(t, err)
	})
}

func TestDBStorage_Get_Counter(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close db storage")
		}
	}()

	storage := &database.DBStorage{DB: db}
	ctx := context.Background()

	rows := sqlmock.NewRows([]string{"MType", "Delta", "Value"}).
		AddRow("counter", int64(42), nil)

	mock.ExpectQuery(`SELECT MType, Delta, Value FROM Metrics WHERE ID = \$1`).
		WithArgs("requests").
		WillReturnRows(rows)

	metric, err := storage.Get(ctx, "counter", "requests")
	require.NoError(t, err)
	require.Equal(t, "requests", metric.ID)
	require.Equal(t, "counter", metric.MType)
	require.Equal(t, int64(42), metric.Delta)
}

func TestDBStorage_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close db storage")
		}
	}()

	storage := &database.DBStorage{DB: db}
	ctx := context.Background()

	t.Run("correct-case-gauge#1", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO Metrics`).
			WithArgs("cpu", "gauge", float64(3.14)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		m := metrics.Metric{
			ID:    "cpu",
			MType: metrics.GaugeName,
			Value: float64(3.14),
		}

		err = storage.Update(ctx, metrics.Metrics{m})
		require.NoError(t, err)
	})

	t.Run("correct-case-counter#1", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(`INSERT INTO Metrics`).
			WithArgs("requests", "counter", int64(5)).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		m := metrics.Metric{
			ID:    "requests",
			MType: metrics.CounterName,
			Delta: int64(5),
		}

		err = storage.Update(ctx, metrics.Metrics{m})
		require.NoError(t, err)
	})

	t.Run("transaction-error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

		m := metrics.Metric{
			ID:    "error",
			MType: metrics.GaugeName,
			Value: float64(1),
		}

		err = storage.Update(ctx, metrics.Metrics{m})
		require.Error(t, err)
	})
}

func TestDBStorage_List(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() {
		if err := db.Close(); err != nil {
			log.Info().Err(err).Msg("fialed to close db")
		}
	}()

	storage := &database.DBStorage{DB: db}
	ctx := context.Background()

	t.Run("correct-case#1", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"ID", "MType", "Delta", "Value"}).
			AddRow("cpu", "gauge", nil, float64(1.5)).
			AddRow("requests", "counter", int64(10), nil)

		mock.ExpectQuery(`SELECT ID, MType, Delta, Value FROM Metrics`).
			WillReturnRows(rows)

		metrics := storage.List(ctx)
		require.Len(t, metrics, 2)

		require.Equal(t, "cpu", metrics[0].ID)
		require.Equal(t, "gauge", metrics[0].MType)
		require.Equal(t, float64(1.5), metrics[0].Value)

		require.Equal(t, "requests", metrics[1].ID)
		require.Equal(t, "counter", metrics[1].MType)
		require.Equal(t, int64(10), metrics[1].Delta)
	})

	t.Run("incorrect-case#1", func(t *testing.T) {
		mock.ExpectQuery(`SELECT ID, MType, Delta, Value FROM Metrics`).
			WillReturnError(sql.ErrConnDone)

		metrics := storage.List(ctx)
		require.Empty(t, metrics)
	})
}
