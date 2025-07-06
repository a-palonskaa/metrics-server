package metrics_storage

import (
	"strconv"
)

type Stringer interface {
	String() string
}

type Gauge float64
type Counter int64

func (val Gauge) String() string {
	return strconv.FormatFloat(float64(val), 'f', -1, 64)
}

func (val Counter) String() string {
	return strconv.FormatInt(int64(val), 10)
}

type MetricsStorage struct {
	GaugeMetrics   map[string]Gauge
	CounterMetrics map[string]Counter

	AllowedGaugeNames   []string
	AllowedCounterNames []string
}

var MS = &MetricsStorage{
	GaugeMetrics:   make(map[string]Gauge),
	CounterMetrics: make(map[string]Counter),

	AllowedGaugeNames: []string{"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys", "HeapAlloc", "HeapIdle",
		"HeapInuse", "HeapObjects", "HeapReleased", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
		"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC", "NumGC", "OtherSys",
		"PauseTotalNs", "StackInuse", "StackSys", "Sys", "TotalAlloc", "RandomValue"},

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

func (m *MetricsStorage) AddGauge(name string, val Gauge) {
	if !m.IsGaugeAllowed(name) {
		m.AllowedGaugeNames = append(m.AllowedGaugeNames, name)
	}
	m.GaugeMetrics[name] = val
}

func (m *MetricsStorage) AddCounter(name string, val Counter) {
	if !m.IsCounterAllowed(name) {
		m.AllowedCounterNames = append(m.AllowedCounterNames, name)
	}
	m.CounterMetrics[name] += val
}

func (m *MetricsStorage) GetGaugeValue(name string) (Gauge, bool) {
	if m.IsGaugeAllowed(name) {
		return m.GaugeMetrics[name], true
	}
	return 0, false
}

func (m *MetricsStorage) GetCounterValue(name string) (Counter, bool) {
	if m.IsCounterAllowed(name) {
		return m.CounterMetrics[name], true
	}
	return 0, false
}
