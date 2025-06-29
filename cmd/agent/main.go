package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"time"
)

var pollInterval time.Duration = 2e9
var reportInterval time.Duration = 1e10

type Gauge float64
type Counter int64

type Metrics struct {
	GaugeMetrics   map[string]Gauge
	CounterMetrics map[string]Counter
}

func update(metrics *Metrics, memStats *runtime.MemStats) {
	runtime.ReadMemStats(memStats)

	// gauge metrics
	metrics.GaugeMetrics["Alloc"] = Gauge(memStats.Alloc)
	metrics.GaugeMetrics["BuckHashSys"] = Gauge(memStats.BuckHashSys)
	metrics.GaugeMetrics["Frees"] = Gauge(memStats.Frees)
	metrics.GaugeMetrics["GCCPUFraction"] = Gauge(memStats.GCCPUFraction)
	metrics.GaugeMetrics["GCSys"] = Gauge(memStats.GCSys)
	metrics.GaugeMetrics["HeapAlloc"] = Gauge(memStats.HeapAlloc)
	metrics.GaugeMetrics["HeapIdle"] = Gauge(memStats.HeapIdle)
	metrics.GaugeMetrics["HeapInuse"] = Gauge(memStats.HeapInuse)
	metrics.GaugeMetrics["HeapObjects"] = Gauge(memStats.HeapObjects)
	metrics.GaugeMetrics["HeapReleased"] = Gauge(memStats.HeapReleased)
	metrics.GaugeMetrics["LastGC"] = Gauge(memStats.LastGC)
	metrics.GaugeMetrics["Lookups"] = Gauge(memStats.Lookups)
	metrics.GaugeMetrics["MCacheInuse"] = Gauge(memStats.MCacheInuse)
	metrics.GaugeMetrics["MCacheSys"] = Gauge(memStats.MCacheSys)
	metrics.GaugeMetrics["MSpanInuse"] = Gauge(memStats.MSpanInuse)
	metrics.GaugeMetrics["MSpanSys"] = Gauge(memStats.MSpanSys)
	metrics.GaugeMetrics["Mallocs"] = Gauge(memStats.Mallocs)
	metrics.GaugeMetrics["NextGC"] = Gauge(memStats.NextGC)
	metrics.GaugeMetrics["NumForcedGC"] = Gauge(memStats.NumForcedGC)
	metrics.GaugeMetrics["NumGC"] = Gauge(memStats.NumGC)
	metrics.GaugeMetrics["OtherSys"] = Gauge(memStats.OtherSys)
	metrics.GaugeMetrics["PauseTotalNs"] = Gauge(memStats.PauseTotalNs)
	metrics.GaugeMetrics["StackInuse"] = Gauge(memStats.StackInuse)
	metrics.GaugeMetrics["StackSys"] = Gauge(memStats.StackSys)
	metrics.GaugeMetrics["Sys"] = Gauge(memStats.Sys)
	metrics.GaugeMetrics["TotalAlloc"] = Gauge(memStats.TotalAlloc)
	metrics.GaugeMetrics["RandomValue"] = Gauge(rand.Float64())

	// counter metrics
	metrics.CounterMetrics["PollCount"]++
}

func updateRoutine(metrics *Metrics, memStats *runtime.MemStats) {
	for {
		time.Sleep(pollInterval)
		update(metrics, memStats)
	}
}

func sendRequest(client *http.Client, kind string, name string, val interface{}) error {
	url := fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", kind, name, val)
	response, err := client.Post(url, "text/html", nil)
	if err != nil {
		fmt.Println(err)
		return err
	}

	_, err = io.Copy(io.Discard, response.Body)
	response.Body.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func main() {
	memStats := &runtime.MemStats{}
	metrics := &Metrics{make(map[string]Gauge, 3), make(map[string]Counter, 1)}

	update(metrics, memStats)
	go updateRoutine(metrics, memStats)

	client := &http.Client{}
	for {
		for key, val := range metrics.GaugeMetrics {
			if err := sendRequest(client, "gauge", key, val); err != nil {
				return
			}
		}

		for key, val := range metrics.CounterMetrics {
			if err := sendRequest(client, "counter", key, val); err != nil {
				return
			}
		}
		time.Sleep(reportInterval)
	}
}
