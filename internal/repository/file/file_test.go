package file_test

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/rs/zerolog/log"

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
