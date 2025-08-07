// Package usecase provides an abstraction over the metrics storage,
// including logic for retrieving, updating, and validating metric data.
package usecase

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

// MemStorage is a usecase layer for metrics collection and storage.
type MemStorage struct {
	storage MetricsRepository
}

// NewMemStorage creates a new MemStorage instance with the given MetricsRepository.
func NewMemStorage(storage MetricsRepository) MemStorage {
	ms := MemStorage{
		storage: storage,
	}
	return ms
}

// GetAllMetrics return all stored metrics from the repository.
func (ms MemStorage) GetAllMetrics(ctx context.Context) []metrics.Metric {
	return ms.storage.List(ctx)
}

// UpdateMetrics updates a list of metrics in the repository.
// Returns an error if any metrics fail to be saved.
func (ms MemStorage) UpdateMetrics(ctx context.Context, mt metrics.Metrics) error {
	if err := ms.storage.Update(ctx, mt); err != nil {
		log.Error().Err(err).Msg("failed to add metric")
		return fmt.Errorf("failed to update metric:%w", err)
	}
	return nil
}

// GetMetric return a single metric by its type and ID from the repository.
// Returns an error if the metric is not found or the type is invalid.
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
