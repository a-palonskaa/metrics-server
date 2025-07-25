package metricsstorage_test

import (
	"context"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
)

func TestMetricsStorage_AddAndGet(t *testing.T) {
	type testCase struct {
		name        string
		beforeAdd   bool
		metric      metrics.Metric
		expectError error
		expectDelta int64
		expectValue float64
	}

	ctx := context.Background()
	storage := memstorage.New()

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

func TestMetricsStorage_List(t *testing.T) {
	ctx := context.Background()
	storage := memstorage.New()

	metricsToAdd := []metrics.Metric{
		{
			ID:    "cpu",
			MType: "gauge",
			Value: 1.5,
		},
		{
			ID:    "reqs",
			MType: "counter",
			Delta: 100,
		},
		{
			ID:    "mem",
			MType: "gauge",
			Value: 256,
		},
	}

	err := storage.Update(ctx, metrics.Metrics(metricsToAdd))
	require.NoError(t, err)

	got := storage.List(ctx)
	require.Len(t, got, len(metricsToAdd))

	sort.Slice(got, func(i, j int) bool { return got[i].ID < got[j].ID })
	sort.Slice(metricsToAdd, func(i, j int) bool { return metricsToAdd[i].ID < metricsToAdd[j].ID })
	assert.Equal(t, metricsToAdd, got)
}

func TestMetricsStorage_Close(t *testing.T) {
	storage := memstorage.New()
	err := storage.Close()
	require.NoError(t, err)
}
