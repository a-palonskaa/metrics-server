package metricsstorage

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	t "github.com/a-palonskaa/metrics-server/internal/metrics"
)

type MetricsStorage struct {
	GaugeMetrics   map[string]t.Gauge
	CounterMetrics map[string]t.Counter

	AllowedGaugeNames   []string
	AllowedCounterNames []string
}

var MS = &MetricsStorage{
	GaugeMetrics:   make(map[string]t.Gauge),
	CounterMetrics: make(map[string]t.Counter),

	AllowedGaugeNames: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
		"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue", "HeapSys"},
	AllowedCounterNames: []string{"PollCount"},
}

func (m *MetricsStorage) IsGaugeAllowed(name string) bool {
	for _, allowed := range m.AllowedGaugeNames {
		if name == allowed {
			return true
		}
	}
	return false
}

func (m *MetricsStorage) IsCounterAllowed(name string) bool {
	for _, allowed := range m.AllowedCounterNames {
		if name == allowed {
			return true
		}
	}
	return false
}

func (m *MetricsStorage) IsNameAllowed(mType, name string) bool {
	switch mType {
	case t.GaugeName:
		return m.IsGaugeAllowed(name)
	case t.CounterName:
		return m.IsCounterAllowed(name)
	}
	return false
}

func (m *MetricsStorage) AddGauge(name string, val t.Gauge) {
	if !m.IsGaugeAllowed(name) {
		m.AllowedGaugeNames = append(m.AllowedGaugeNames, name)
	}
	m.GaugeMetrics[name] = val
}

func (m *MetricsStorage) AddCounter(name string, val t.Counter) {
	if !m.IsCounterAllowed(name) {
		m.AllowedCounterNames = append(m.AllowedCounterNames, name)
	}
	m.CounterMetrics[name] += val
}

func (m *MetricsStorage) AddValue(mType, name string, val any) bool {
	switch mType {
	case t.GaugeName:
		if v, ok := val.(t.Gauge); ok {
			m.AddGauge(name, v)
			return true
		}
	case t.CounterName:
		if v, ok := val.(t.Counter); ok {
			m.AddCounter(name, v)
			return true
		}
	}
	return false
}

func (m *MetricsStorage) GetGaugeValue(name string) (t.Gauge, bool) {
	if m.IsGaugeAllowed(name) {
		return m.GaugeMetrics[name], true
	}
	return 0, false
}

func (m *MetricsStorage) GetCounterValue(name string) (t.Counter, bool) {
	if m.IsCounterAllowed(name) {
		return m.CounterMetrics[name], true
	}
	return 0, false
}

func (m *MetricsStorage) GetValue(mType, name string) (any, bool) {
	switch mType {
	case t.GaugeName:
		val, ok := m.GetGaugeValue(name)
		return val, ok
	case t.CounterName:
		val, ok := m.GetCounterValue(name)
		return val, ok
	default:
		return nil, false
	}
}

func IsTypeAllowed(mType string) bool {
	return mType == t.GaugeName || mType == t.CounterName
}

func (m *MetricsStorage) Update(memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	// gauge metrics
	m.GaugeMetrics["Alloc"] = t.Gauge(memStats.Alloc)
	m.GaugeMetrics["BuckHashSys"] = t.Gauge(memStats.BuckHashSys)
	m.GaugeMetrics["Frees"] = t.Gauge(memStats.Frees)
	m.GaugeMetrics["GCCPUFraction"] = t.Gauge(memStats.GCCPUFraction)
	m.GaugeMetrics["GCSys"] = t.Gauge(memStats.GCSys)
	m.GaugeMetrics["HeapAlloc"] = t.Gauge(memStats.HeapAlloc)
	m.GaugeMetrics["HeapIdle"] = t.Gauge(memStats.HeapIdle)
	m.GaugeMetrics["HeapInuse"] = t.Gauge(memStats.HeapInuse)
	m.GaugeMetrics["HeapObjects"] = t.Gauge(memStats.HeapObjects)
	m.GaugeMetrics["HeapReleased"] = t.Gauge(memStats.HeapReleased)
	m.GaugeMetrics["LastGC"] = t.Gauge(memStats.LastGC)
	m.GaugeMetrics["Lookups"] = t.Gauge(memStats.Lookups)
	m.GaugeMetrics["MCacheInuse"] = t.Gauge(memStats.MCacheInuse)
	m.GaugeMetrics["MCacheSys"] = t.Gauge(memStats.MCacheSys)
	m.GaugeMetrics["MSpanInuse"] = t.Gauge(memStats.MSpanInuse)
	m.GaugeMetrics["MSpanSys"] = t.Gauge(memStats.MSpanSys)
	m.GaugeMetrics["Mallocs"] = t.Gauge(memStats.Mallocs)
	m.GaugeMetrics["NextGC"] = t.Gauge(memStats.NextGC)
	m.GaugeMetrics["NumForcedGC"] = t.Gauge(memStats.NumForcedGC)
	m.GaugeMetrics["NumGC"] = t.Gauge(memStats.NumGC)
	m.GaugeMetrics["OtherSys"] = t.Gauge(memStats.OtherSys)
	m.GaugeMetrics["PauseTotalNs"] = t.Gauge(memStats.PauseTotalNs)
	m.GaugeMetrics["StackInuse"] = t.Gauge(memStats.StackInuse)
	m.GaugeMetrics["StackSys"] = t.Gauge(memStats.StackSys)
	m.GaugeMetrics["Sys"] = t.Gauge(memStats.Sys)
	m.GaugeMetrics["TotalAlloc"] = t.Gauge(memStats.TotalAlloc)
	m.GaugeMetrics["HeapSys"] = t.Gauge(memStats.HeapSys)
	m.GaugeMetrics["RandomValue"] = t.Gauge(rand.Float64())

	// counter metrics
	m.CounterMetrics["PollCount"]++
}

func (m *MetricsStorage) UpdateRoutine(memStats *runtime.MemStats, interval time.Duration) {
	for {
		time.Sleep(interval)
		m.Update(memStats)
	}
}

func (m *MetricsStorage) Iterate(f func(string, string, fmt.Stringer)) {
	for key, value := range m.GaugeMetrics {
		f(key, t.GaugeName, value)
	}

	for key, value := range m.CounterMetrics {
		f(key, t.CounterName, value)
	}
}
