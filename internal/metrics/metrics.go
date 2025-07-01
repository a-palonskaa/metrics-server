package metrics

import (
	"math/rand"
	"runtime"
	"time"

	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

var PollInterval time.Duration = 2e9
var ReportInterval time.Duration = 1e10

func Update(metrics *st.MetricsStorage, memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	// gauge metrics
	metrics.GaugeMetrics["Alloc"] = st.Gauge(memStats.Alloc)
	metrics.GaugeMetrics["BuckHashSys"] = st.Gauge(memStats.BuckHashSys)
	metrics.GaugeMetrics["Frees"] = st.Gauge(memStats.Frees)
	metrics.GaugeMetrics["GCCPUFraction"] = st.Gauge(memStats.GCCPUFraction)
	metrics.GaugeMetrics["GCSys"] = st.Gauge(memStats.GCSys)
	metrics.GaugeMetrics["HeapAlloc"] = st.Gauge(memStats.HeapAlloc)
	metrics.GaugeMetrics["HeapIdle"] = st.Gauge(memStats.HeapIdle)
	metrics.GaugeMetrics["HeapInuse"] = st.Gauge(memStats.HeapInuse)
	metrics.GaugeMetrics["HeapObjects"] = st.Gauge(memStats.HeapObjects)
	metrics.GaugeMetrics["HeapReleased"] = st.Gauge(memStats.HeapReleased)
	metrics.GaugeMetrics["LastGC"] = st.Gauge(memStats.LastGC)
	metrics.GaugeMetrics["Lookups"] = st.Gauge(memStats.Lookups)
	metrics.GaugeMetrics["MCacheInuse"] = st.Gauge(memStats.MCacheInuse)
	metrics.GaugeMetrics["MCacheSys"] = st.Gauge(memStats.MCacheSys)
	metrics.GaugeMetrics["MSpanInuse"] = st.Gauge(memStats.MSpanInuse)
	metrics.GaugeMetrics["MSpanSys"] = st.Gauge(memStats.MSpanSys)
	metrics.GaugeMetrics["Mallocs"] = st.Gauge(memStats.Mallocs)
	metrics.GaugeMetrics["NextGC"] = st.Gauge(memStats.NextGC)
	metrics.GaugeMetrics["NumForcedGC"] = st.Gauge(memStats.NumForcedGC)
	metrics.GaugeMetrics["NumGC"] = st.Gauge(memStats.NumGC)
	metrics.GaugeMetrics["OtherSys"] = st.Gauge(memStats.OtherSys)
	metrics.GaugeMetrics["PauseTotalNs"] = st.Gauge(memStats.PauseTotalNs)
	metrics.GaugeMetrics["StackInuse"] = st.Gauge(memStats.StackInuse)
	metrics.GaugeMetrics["StackSys"] = st.Gauge(memStats.StackSys)
	metrics.GaugeMetrics["Sys"] = st.Gauge(memStats.Sys)
	metrics.GaugeMetrics["TotalAlloc"] = st.Gauge(memStats.TotalAlloc)
	metrics.GaugeMetrics["RandomValue"] = st.Gauge(rand.Float64())

	// counter metrics
	metrics.CounterMetrics["PollCount"]++
}

func UpdateRoutine(metrics *st.MetricsStorage, memStats *runtime.MemStats) {
	for {
		time.Sleep(PollInterval)
		Update(metrics, memStats)
	}
}
