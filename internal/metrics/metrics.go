package metrics

import (
	"math/rand"
	"runtime"
	"time"

	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

var PollInterval int = 2
var ReportInterval int = 10

func Update(metrics *memstorage.MetricsStorage, memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	// gauge metrics
	metrics.GaugeMetrics["Alloc"] = memstorage.Gauge(memStats.Alloc)
	metrics.GaugeMetrics["BuckHashSys"] = memstorage.Gauge(memStats.BuckHashSys)
	metrics.GaugeMetrics["Frees"] = memstorage.Gauge(memStats.Frees)
	metrics.GaugeMetrics["GCCPUFraction"] = memstorage.Gauge(memStats.GCCPUFraction)
	metrics.GaugeMetrics["GCSys"] = memstorage.Gauge(memStats.GCSys)
	metrics.GaugeMetrics["HeapAlloc"] = memstorage.Gauge(memStats.HeapAlloc)
	metrics.GaugeMetrics["HeapIdle"] = memstorage.Gauge(memStats.HeapIdle)
	metrics.GaugeMetrics["HeapInuse"] = memstorage.Gauge(memStats.HeapInuse)
	metrics.GaugeMetrics["HeapObjects"] = memstorage.Gauge(memStats.HeapObjects)
	metrics.GaugeMetrics["HeapReleased"] = memstorage.Gauge(memStats.HeapReleased)
	metrics.GaugeMetrics["LastGC"] = memstorage.Gauge(memStats.LastGC)
	metrics.GaugeMetrics["Lookups"] = memstorage.Gauge(memStats.Lookups)
	metrics.GaugeMetrics["MCacheInuse"] = memstorage.Gauge(memStats.MCacheInuse)
	metrics.GaugeMetrics["MCacheSys"] = memstorage.Gauge(memStats.MCacheSys)
	metrics.GaugeMetrics["MSpanInuse"] = memstorage.Gauge(memStats.MSpanInuse)
	metrics.GaugeMetrics["MSpanSys"] = memstorage.Gauge(memStats.MSpanSys)
	metrics.GaugeMetrics["Mallocs"] = memstorage.Gauge(memStats.Mallocs)
	metrics.GaugeMetrics["NextGC"] = memstorage.Gauge(memStats.NextGC)
	metrics.GaugeMetrics["NumForcedGC"] = memstorage.Gauge(memStats.NumForcedGC)
	metrics.GaugeMetrics["NumGC"] = memstorage.Gauge(memStats.NumGC)
	metrics.GaugeMetrics["OtherSys"] = memstorage.Gauge(memStats.OtherSys)
	metrics.GaugeMetrics["PauseTotalNs"] = memstorage.Gauge(memStats.PauseTotalNs)
	metrics.GaugeMetrics["StackInuse"] = memstorage.Gauge(memStats.StackInuse)
	metrics.GaugeMetrics["StackSys"] = memstorage.Gauge(memStats.StackSys)
	metrics.GaugeMetrics["Sys"] = memstorage.Gauge(memStats.Sys)
	metrics.GaugeMetrics["TotalAlloc"] = memstorage.Gauge(memStats.TotalAlloc)
	metrics.GaugeMetrics["HeapSys"] = memstorage.Gauge(memStats.HeapSys)
	metrics.GaugeMetrics["RandomValue"] = memstorage.Gauge(rand.Float64())

	// counter metrics
	metrics.CounterMetrics["PollCount"]++
}

func UpdateRoutine(metrics *memstorage.MetricsStorage, memStats *runtime.MemStats) {
	for {
		time.Sleep(time.Duration(PollInterval) * 1e9)
		Update(metrics, memStats)
	}
}
