package main

import (
	"net/http"
	"runtime"
	"time"

	ha "github.com/a-palonskaa/metrics-server/internal/handlers/agent"
	mt "github.com/a-palonskaa/metrics-server/internal/metrics"
	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func main() {
	memStats := &runtime.MemStats{}

	mt.Update(st.MS, memStats)
	go mt.UpdateRoutine(st.MS, memStats)

	client := &http.Client{}
	for {
		for key, val := range st.MS.GaugeMetrics {
			if err := ha.SendRequest(client, "gauge", key, val); err != nil {
				return
			}
		}

		for key, val := range st.MS.CounterMetrics {
			if err := ha.SendRequest(client, "counter", key, val); err != nil {
				return
			}
		}
		time.Sleep(mt.ReportInterval)
	}
}
