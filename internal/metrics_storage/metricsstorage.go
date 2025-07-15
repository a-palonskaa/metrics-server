package metricsstorage

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
)

type MemStorage interface {
	IsGaugeAllowed(name string) bool
	IsCounterAllowed(name string) bool
	IsNameAllowed(mType, name string) bool
	AddGauge(name string, val metrics.Gauge)
	AddCounter(name string, val metrics.Counter)
	GetGaugeValue(name string) (metrics.Gauge, bool)
	GetCounterValue(name string) (metrics.Counter, bool)
	Update(memStats *runtime.MemStats)
	Iterate(f func(string, string, fmt.Stringer))
	AddMetricsToStorage(metrics *metrics.MetricsS) int
}

//easyjson:json
type MetricsStorage struct {
	GaugeMetrics   map[string]metrics.Gauge
	CounterMetrics map[string]metrics.Counter

	AllowedGaugeNames   map[string]bool
	AllowedCounterNames map[string]bool
}

var MS = &MetricsStorage{
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

func (m *MetricsStorage) IsGaugeAllowed(name string) bool {
	return m.AllowedGaugeNames[name]
}

func (m *MetricsStorage) IsCounterAllowed(name string) bool {
	return m.AllowedCounterNames[name]
}

func (m *MetricsStorage) IsNameAllowed(mType, name string) bool {
	switch mType {
	case metrics.GaugeName:
		return m.IsGaugeAllowed(name)
	case metrics.CounterName:
		return m.IsCounterAllowed(name)
	}
	return false
}

func (m *MetricsStorage) AddGauge(name string, val metrics.Gauge) {
	if !m.IsGaugeAllowed(name) {
		m.AllowedGaugeNames[name] = true
	}
	m.GaugeMetrics[name] = val
}

func (m *MetricsStorage) AddCounter(name string, val metrics.Counter) {
	if !m.IsCounterAllowed(name) {
		m.AllowedCounterNames[name] = true
	}
	m.CounterMetrics[name] += val
}

func (m *MetricsStorage) AddValue(mType, name string, val any) bool {
	switch mType {
	case metrics.GaugeName:
		if v, ok := val.(metrics.Gauge); ok {
			m.AddGauge(name, v)
			return true
		}
	case metrics.CounterName:
		if v, ok := val.(metrics.Counter); ok {
			m.AddCounter(name, v)
			return true
		}
	}
	return false
}

func (m *MetricsStorage) GetGaugeValue(name string) (metrics.Gauge, bool) {
	if m.IsGaugeAllowed(name) {
		return m.GaugeMetrics[name], true
	}
	return 0, false
}

func (m *MetricsStorage) GetCounterValue(name string) (metrics.Counter, bool) {
	if m.IsCounterAllowed(name) {
		return m.CounterMetrics[name], true
	}
	return 0, false
}

func (m *MetricsStorage) GetValue(mType, name string) (any, bool) {
	switch mType {
	case metrics.GaugeName:
		val, ok := m.GetGaugeValue(name)
		return val, ok
	case metrics.CounterName:
		val, ok := m.GetCounterValue(name)
		return val, ok
	default:
		return nil, false
	}
}

func IsTypeAllowed(mType string) bool {
	return mType == metrics.GaugeName || mType == metrics.CounterName
}

func (m *MetricsStorage) Update(memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	// gauge metrics
	m.GaugeMetrics["Alloc"] = metrics.Gauge(memStats.Alloc)
	m.GaugeMetrics["BuckHashSys"] = metrics.Gauge(memStats.BuckHashSys)
	m.GaugeMetrics["Frees"] = metrics.Gauge(memStats.Frees)
	m.GaugeMetrics["GCCPUFraction"] = metrics.Gauge(memStats.GCCPUFraction)
	m.GaugeMetrics["GCSys"] = metrics.Gauge(memStats.GCSys)
	m.GaugeMetrics["HeapAlloc"] = metrics.Gauge(memStats.HeapAlloc)
	m.GaugeMetrics["HeapIdle"] = metrics.Gauge(memStats.HeapIdle)
	m.GaugeMetrics["HeapInuse"] = metrics.Gauge(memStats.HeapInuse)
	m.GaugeMetrics["HeapObjects"] = metrics.Gauge(memStats.HeapObjects)
	m.GaugeMetrics["HeapReleased"] = metrics.Gauge(memStats.HeapReleased)
	m.GaugeMetrics["LastGC"] = metrics.Gauge(memStats.LastGC)
	m.GaugeMetrics["Lookups"] = metrics.Gauge(memStats.Lookups)
	m.GaugeMetrics["MCacheInuse"] = metrics.Gauge(memStats.MCacheInuse)
	m.GaugeMetrics["MCacheSys"] = metrics.Gauge(memStats.MCacheSys)
	m.GaugeMetrics["MSpanInuse"] = metrics.Gauge(memStats.MSpanInuse)
	m.GaugeMetrics["MSpanSys"] = metrics.Gauge(memStats.MSpanSys)
	m.GaugeMetrics["Mallocs"] = metrics.Gauge(memStats.Mallocs)
	m.GaugeMetrics["NextGC"] = metrics.Gauge(memStats.NextGC)
	m.GaugeMetrics["NumForcedGC"] = metrics.Gauge(memStats.NumForcedGC)
	m.GaugeMetrics["NumGC"] = metrics.Gauge(memStats.NumGC)
	m.GaugeMetrics["OtherSys"] = metrics.Gauge(memStats.OtherSys)
	m.GaugeMetrics["PauseTotalNs"] = metrics.Gauge(memStats.PauseTotalNs)
	m.GaugeMetrics["StackInuse"] = metrics.Gauge(memStats.StackInuse)
	m.GaugeMetrics["StackSys"] = metrics.Gauge(memStats.StackSys)
	m.GaugeMetrics["Sys"] = metrics.Gauge(memStats.Sys)
	m.GaugeMetrics["TotalAlloc"] = metrics.Gauge(memStats.TotalAlloc)
	m.GaugeMetrics["HeapSys"] = metrics.Gauge(memStats.HeapSys)
	m.GaugeMetrics["RandomValue"] = metrics.Gauge(rand.Float64())

	// counter metrics
	m.CounterMetrics["PollCount"]++
}

func (m *MetricsStorage) Iterate(f func(string, string, fmt.Stringer)) {
	for key, value := range m.GaugeMetrics {
		f(key, metrics.GaugeName, value)
	}

	for key, value := range m.CounterMetrics {
		f(key, metrics.CounterName, value)
	}
}

func (m *MetricsStorage) AddMetricsToStorage(mt *metrics.MetricsS) int {
	for _, metric := range *mt {
		switch metric.MType {
		case "gauge":
			m.AddGauge(metric.ID, metrics.Gauge(*metric.Value))
		case "counter":
			m.AddCounter(metric.ID, metrics.Counter(*metric.Delta))
		default:
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}
