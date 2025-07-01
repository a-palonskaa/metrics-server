package main

import (
	"net/http"
	"runtime"
	"time"

	ha "github.com/a-palonskaa/metrics-server/internal/handlers/agent"
	mt "github.com/a-palonskaa/metrics-server/internal/metrics"
	st "github.com/a-palonskaa/metrics-server/internal/storage"
)

func main() {
	memStats := &runtime.MemStats{}
	metrics := &mt.MetricsStorage{make(map[string]st.Gauge, 3), make(map[string]st.Counter, 1)}

	mt.Update(metrics, memStats)
	go mt.UpdateRoutine(metrics, memStats)

	client := &http.Client{}
	for {
		for key, val := range metrics.GaugeMetrics {
			if err := ha.SendRequest(client, "gauge", key, val); err != nil {
				return
			}
		}

		for key, val := range metrics.CounterMetrics {
			if err := ha.SendRequest(client, "counter", key, val); err != nil {
				return
			}
		}
		time.Sleep(mt.ReportInterval)
	}
}
