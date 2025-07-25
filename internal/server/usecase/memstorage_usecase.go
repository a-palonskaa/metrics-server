package usecase

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

type MemStorage struct {
	storage MetricsRepository
}

func NewMemStorage(storage MetricsRepository) MemStorage {
	ms := MemStorage{
		storage: storage,
	}
	return ms
}

func (ms MemStorage) GetAllMetrics(ctx context.Context) []metrics.Metric {
	return ms.storage.List(ctx)
}

func (ms MemStorage) UpdateMetrics(ctx context.Context, mt metrics.Metrics) error {
	if err := ms.storage.Update(ctx, mt); err != nil {
		log.Error().Err(err).Msg("failed to add metric")
		return err
	}
	return nil
}

func (ms MemStorage) GetMetric(ctx context.Context, metric *metrics.Metric) error {
	var err error
	if !metrics.IsTypeAllowed(metric.MType) {
		log.Error().Msgf("invalid type %sd", metric.MType)
		return fmt.Errorf("%w: %s", metrics.ErrIncorrectMetricType, metric.MType)
	}
	*metric, err = ms.storage.Get(ctx, metric.MType, metric.ID)
	if err != nil {
		log.Error().Err(err).Msgf("metric %s not found", metric.ID)
		return fmt.Errorf("%w: %s", metrics.ErrUnallowedMetric, metric.ID)
	}
	return nil
}
