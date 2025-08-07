// Package metricsstorage provides an in-memory implementation of the MetricsRepository interface.
package metricsstorage

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

// MetricsStorage represents an in-memory storage for metrics.
type MetricsStorage struct {
	mu      sync.RWMutex
	metrics map[string]map[string]metrics.Metric
}

// New creates and returns a new MetricsStorage instance.
func New() *MetricsStorage {
	return &MetricsStorage{
		metrics: make(map[string]map[string]metrics.Metric),
	}
}

func (m *MetricsStorage) Update(ctx context.Context, mt metrics.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gauges, ok := m.metrics[metrics.GaugeName]
	if !ok {
		gauges = make(map[string]metrics.Metric, len(mt))
		m.metrics[metrics.GaugeName] = gauges
	}

	counters, ok := m.metrics[metrics.CounterName]
	if !ok {
		counters = make(map[string]metrics.Metric, len(mt))
		m.metrics[metrics.CounterName] = counters
	}

	for i := range mt {
		if !metrics.IsTypeAllowed(mt[i].MType) {
			log.Error().Str("type", mt[i].MType).Msg("unallowed metric type")
			return metrics.ErrIncorrectMetricType
		}

		id := mt[i].ID
		switch mt[i].MType {
		case metrics.GaugeName:
			gauges[id] = mt[i]
		case metrics.CounterName:
			if existing, ok := counters[id]; ok {
				mt[i].Delta += existing.Delta
			}
			counters[id] = mt[i]
		}
	}
	return nil
}

func (m *MetricsStorage) Get(ctx context.Context, mType, name string) (metrics.Metric, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !metrics.IsTypeAllowed(mType) {
		log.Error().Msgf("unallowed type %s", mType)
		return metrics.Metric{}, metrics.ErrIncorrectMetricType
	}

	metric, ok := m.metrics[mType][name]
	if !ok {
		log.Error().Msgf("unallowed name %s of type %s", name, mType)
		return metrics.Metric{}, metrics.ErrUnallowedMetric
	}
	return metric, nil
}

func (m *MetricsStorage) List(ctx context.Context) []metrics.Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var mt []metrics.Metric
	for _, metricsMap := range m.metrics {
		for _, metric := range metricsMap {
			mt = append(mt, metric)
		}
	}
	return mt
}

func (m *MetricsStorage) Close() error {
	return nil
}
