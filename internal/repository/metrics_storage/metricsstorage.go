package metricsstorage

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

// ----------------------MetricsStorage-type----------------------
type MetricsStorage struct {
	GaugeMetrics   map[string]metrics.Gauge
	CounterMetrics map[string]metrics.Counter

	AllowedGaugeNames   map[string]bool
	AllowedCounterNames map[string]bool
}

func NewMetricsStorage() *MetricsStorage {
	return &MetricsStorage{
		GaugeMetrics:   make(map[string]metrics.Gauge),
		CounterMetrics: make(map[string]metrics.Counter),

		AllowedGaugeNames: map[string]bool{
			"Alloc": true, "BuckHashSys": true, "Frees": true, "GCCPUFraction": true, "GCSys": true,
			"HeapAlloc": true, "HeapIdle": true, "HeapInuse": true, "HeapObjects": true, "HeapReleased": true,
			"LastGC": true, "Lookups": true, "MCacheInuse": true, "MCacheSys": true, "MSpanInuse": true,
			"MSpanSys": true, "Mallocs": true, "NextGC": true, "NumForcedGC": true, "NumGC": true, "OtherSys": true,
			"PauseTotalNs": true, "StackInuse": true, "StackSys": true, "Sys": true, "TotalAlloc": true,
			"RandomValue": true, "HeapSys": true},
		AllowedCounterNames: map[string]bool{"PollCount": true},
	}
}

// ----------------------MemStorageInterface----------------------
func (m *MetricsStorage) Add(_ context.Context, mType, name string, val fmt.Stringer) {
	switch mType {
	case metrics.GaugeName:
		if v, ok := val.(metrics.Gauge); ok {
			m.AddGauge(context.TODO(), name, v)
		}
	case metrics.CounterName:
		if v, ok := val.(metrics.Counter); ok {
			m.AddCounter(context.TODO(), name, v)
		}
	default:
		log.Error().Msgf("unallowed type %s", mType)
	}
}

func (m *MetricsStorage) Get(_ context.Context, mType, name string) (fmt.Stringer, bool) {
	if ok := m.IsNameAllowed(context.TODO(), mType, name); !ok {
		return nil, false
	}

	switch mType {
	case metrics.GaugeName:
		val, ok := m.GetGaugeValue(context.TODO(), name)
		return val, ok
	case metrics.CounterName:
		val, ok := m.GetCounterValue(context.TODO(), name)
		return val, ok
	}
	return nil, false
}

func (m *MetricsStorage) List(ctx context.Context) []metrics.Metric {
	allMetrics := make([]metrics.Metric, 0, len(m.AllowedGaugeNames)+len(m.AllowedCounterNames))
	for key, value := range m.GaugeMetrics {
		val := float64(value)
		allMetrics = append(allMetrics, metrics.Metric{
			ID:    key,
			MType: metrics.GaugeName,
			Value: &val,
		})
	}

	for key, value := range m.CounterMetrics {
		val := int64(value)
		allMetrics = append(allMetrics, metrics.Metric{
			ID:    key,
			MType: metrics.CounterName,
			Delta: &val,
		})
	}
	return allMetrics
}

func (m *MetricsStorage) Close() error {
	return nil
}

// ----------------------MemStorage-methods----------------------
// ------------------is-allowed------------------
func (m *MetricsStorage) IsNameAllowed(_ context.Context, mType, name string) bool {
	switch mType {
	case metrics.GaugeName:
		return m.IsGaugeAllowed(context.TODO(), name)
	case metrics.CounterName:
		return m.IsCounterAllowed(context.TODO(), name)
	}
	return false
}

func (m *MetricsStorage) IsGaugeAllowed(_ context.Context, name string) bool {
	return m.AllowedGaugeNames[name]
}

func (m *MetricsStorage) IsCounterAllowed(_ context.Context, name string) bool {
	return m.AllowedCounterNames[name]
}

// ------------------add------------------
func (m *MetricsStorage) AddGauge(_ context.Context, name string, val metrics.Gauge) {
	if !m.IsGaugeAllowed(context.TODO(), name) {
		m.AllowedGaugeNames[name] = true
	}
	m.GaugeMetrics[name] = val
}

func (m *MetricsStorage) AddCounter(_ context.Context, name string, val metrics.Counter) {
	if !m.IsCounterAllowed(context.TODO(), name) {
		m.AllowedCounterNames[name] = true
	}
	m.CounterMetrics[name] += val
}

func (m *MetricsStorage) GetGaugeValue(_ context.Context, name string) (metrics.Gauge, bool) {
	if m.IsGaugeAllowed(context.TODO(), name) {
		return m.GaugeMetrics[name], true
	}
	return 0, false
}

func (m *MetricsStorage) GetCounterValue(_ context.Context, name string) (metrics.Counter, bool) {
	if m.IsCounterAllowed(context.TODO(), name) {
		return m.CounterMetrics[name], true
	}
	return 0, false
}
