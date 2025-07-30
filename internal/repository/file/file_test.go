package file_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	file "github.com/a-palonskaa/metrics-server/internal/repository/file"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
)

func BenchmarkFileStorage_Save(b *testing.B) {
	_, err := os.CreateTemp("", "metrics.json")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := os.Remove("metrics.json"); err != nil {
			log.Info().Err(err).Msg("failed to remove metric.json")
		}
	}()

	fs := file.New("metrics.json", memstorage.New(), 0, false)
	defer func() {
		if err := fs.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close filestorage")
		}
	}()
	var gaugeMetrics metrics.Metrics
	var counterMetrics metrics.Metrics

	for i := 0; i < 10; i++ {
		id := strconv.Itoa(i)
		gaugeMetrics = append(gaugeMetrics, metrics.Metric{
			ID:    id,
			MType: "gauge",
			Value: float64(i),
		})
		counterMetrics = append(counterMetrics, metrics.Metric{
			ID:    id,
			MType: "counter",
			Delta: int64(i),
		})
	}

	_ = fs.Update(context.Background(), gaugeMetrics)
	_ = fs.Update(context.Background(), counterMetrics)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.Save(context.Background())
	}
}

func createTempFile(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "test_metrics_*.json")
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Info().Err(err).Msg("failed to remove tmp file")
		}
	})
	return tmpFile.Name()
}

func TestFileStorage_UpdateAndGet(t *testing.T) {
	path := createTempFile(t)
	mem := memstorage.New()

	fs := file.New(path, mem, 0, false)
	defer func() {
		if err := fs.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close file storage")
		}
	}()

	m := metrics.Metrics{
		{
			ID:    "Alloc",
			MType: "gauge",
			Value: float64(123.45),
		},
	}

	err := fs.Update(context.Background(), m)
	require.NoError(t, err)

	got, err := fs.Get(context.Background(), "gauge", "Alloc")
	require.NoError(t, err)
	assert.Equal(t, m[0].Value, got.Value)
}

func TestFileStorage_List(t *testing.T) {
	path := createTempFile(t)
	mem := memstorage.New()
	fs := file.New(path, mem, 0, false)
	defer func() {
		if err := fs.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close file storage")
		}
	}()

	m := metrics.Metrics{
		{
			ID:    "TestMetric",
			MType: "counter",
			Delta: int64(42),
		},
	}
	err := fs.Update(context.Background(), m)
	require.NoError(t, err)

	all := fs.List(context.Background())
	require.Len(t, all, 1)
	assert.Equal(t, "TestMetric", all[0].ID)
}

func TestFileStorageSave(t *testing.T) {
	path := createTempFile(t)
	mem := memstorage.New()
	fs := file.New(path, mem, 0, false)
	defer func() {
		if err := fs.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close file storage")
		}
	}()

	m := metrics.Metrics{
		{
			ID:    "DiskUsage",
			MType: "gauge",
			Value: float64(88.8),
		},
	}
	err := fs.Update(context.Background(), m)
	require.NoError(t, err)
}

func TestFileStorage_Close(t *testing.T) {
	path := createTempFile(t)
	fs := file.New(path, memstorage.New(), 0, false)

	err := fs.Close()
	assert.NoError(t, err)
}

func TestMetricsStorage_AddAndGet(t *testing.T) {
	path := createTempFile(t)
	mem := memstorage.New()
	storage := file.New(path, mem, 0, false)
	defer func() {
		if err := storage.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close storage")
		}
	}()

	type testCase struct {
		name        string
		beforeAdd   bool
		metric      metrics.Metric
		expectError error
		expectDelta int64
		expectValue float64
	}

	ctx := context.Background()

	tests := []testCase{
		{
			name:        "get-missing-gauge",
			beforeAdd:   true,
			metric:      metrics.Metric{ID: "not_found", MType: "gauge"},
			expectError: metrics.ErrUnallowedMetric,
		},
		{
			name:        "get-invalid-type",
			beforeAdd:   true,
			metric:      metrics.Metric{ID: "invalid", MType: "invalid"},
			expectError: metrics.ErrIncorrectMetricType,
		},
		{
			name:        "add-valid-gauge",
			metric:      metrics.Metric{ID: "test_gauge", MType: "gauge", Value: 42.42},
			expectValue: 42.42,
		},
		{
			name:        "add-invalid-type",
			metric:      metrics.Metric{ID: "bad", MType: "invalid", Value: 1.0},
			expectError: metrics.ErrIncorrectMetricType,
		},
		{
			name:        "add-valid-gauge",
			metric:      metrics.Metric{ID: "test_gauge", MType: "gauge", Value: 42.42},
			expectValue: 42.42,
		},
		{
			name:        "add-valid-counter",
			metric:      metrics.Metric{ID: "test_counter", MType: "counter", Delta: int64(10)},
			expectDelta: 10,
		},
		{
			name:        "add-counter-sum",
			metric:      metrics.Metric{ID: "test_counter", MType: "counter", Delta: int64(15)},
			expectDelta: 25,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.beforeAdd {
				_, err := storage.Get(ctx, tc.metric.MType, tc.metric.ID)
				require.ErrorIs(t, err, tc.expectError)
				return
			}

			err := storage.Update(ctx, metrics.Metrics([]metrics.Metric{tc.metric}))
			if tc.expectError != nil {
				require.ErrorIs(t, err, tc.expectError)
				return
			}
			require.NoError(t, err)

			got, err := storage.Get(ctx, tc.metric.MType, tc.metric.ID)
			require.NoError(t, err)

			assert.Equal(t, tc.metric.ID, got.ID)
			assert.Equal(t, tc.metric.MType, got.MType)

			if tc.metric.MType == "gauge" {
				assert.Equal(t, tc.expectValue, got.Value)
			}
			if tc.metric.MType == "counter" {
				require.NotNil(t, got.Delta)
				assert.Equal(t, tc.expectDelta, got.Delta)
			}
		})
	}
}
