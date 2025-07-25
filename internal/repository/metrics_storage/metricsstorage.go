package metricsstorage

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

type MetricsStorage struct {
	mu      sync.RWMutex
	metrics map[string]map[string]metrics.Metric
}

func New() *MetricsStorage {
	return &MetricsStorage{
		metrics: make(map[string]map[string]metrics.Metric),
	}
}

func (m *MetricsStorage) Update(ctx context.Context, mt metrics.Metrics) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, metric := range []metrics.Metric(mt) {
		if !metrics.IsTypeAllowed(metric.MType) {
			log.Error().Msgf("unallowed type %s", metric.MType)
			return metrics.ErrIncorrectMetricType
		}

		if m.metrics[metric.MType] == nil {
			m.metrics[metric.MType] = make(map[string]metrics.Metric)
		}

		mt, ok := m.metrics[metric.MType][metric.ID]
		if !ok {
			m.metrics[metric.MType][metric.ID] = metric
			continue
		}

		switch metric.MType {
		case metrics.GaugeName:
			mt.Value = metric.Value
		case metrics.CounterName:
			mt.Delta += metric.Delta
		default:
			return metrics.ErrIncorrectMetricType
		}
		m.metrics[metric.MType][metric.ID] = mt
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
