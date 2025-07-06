package server

import (
	"bytes"
	"fmt"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	mt "github.com/a-palonskaa/metrics-server/internal/metrics"
	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func PostHandler(w http.ResponseWriter, req *http.Request) {
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	if kind != "gauge" && kind != "counter" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if name == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if val == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	switch kind {
	case "gauge":
		{
			gaugeValue, err := strconv.ParseFloat(val, 64)
			if err != nil {
				http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
				return
			}
			st.MS.AddGauge(name, st.Gauge(gaugeValue))
		}
	case "counter":
		{
			counterValue, err := strconv.Atoi(val)
			if err != nil {
				http.Error(w, "Incorrect couner value", http.StatusBadRequest)
				return
			}
			st.MS.AddCounter(name, st.Counter(counterValue))
		}
	}
	w.WriteHeader(http.StatusOK)
}

func AllValueHandler(w http.ResponseWriter, req *http.Request) {
	segments := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	if len(segments) > 1 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	mt.Update(st.MS, &runtime.MemStats{})

	var buf bytes.Buffer

	buf.WriteString(`
        <html>
        <body>
            <h1>MetricsStorage</h1>
            <h2>Gauge MetricsStorage</h2>
            <table border='1' cellpadding='5' cellspacing='0'>
                <tr><th>Name</th><th>Value</th></tr>
    `)

	for _, key := range st.MS.AllowedGaugeNames {
		fmt.Fprintf(&buf, "<tr><td>%s</td><td>%v</td></tr>\n", key, st.MS.GaugeMetrics[key])
	}

	buf.WriteString(`
            </table>
            <h2>Counter MetricsStorage</h2>
            <table border='1' cellpadding='5' cellspacing='0'>
                <tr><th>Name</th><th>Value</th></tr>
    `)

	for _, key := range st.MS.AllowedCounterNames {
		fmt.Fprintf(&buf, "<tr><td>%s</td><td>%v</td></tr>\n", key, st.MS.CounterMetrics[key])
	}

	buf.WriteString(`
            </table>
            </body>
            </html>
    `)

	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func GetHandler(w http.ResponseWriter, req *http.Request) {
	kind := chi.URLParam(req, "kind")
	name := chi.URLParam(req, "name")

	switch kind {
	case "gauge":
		GaugeGetHandler(w, req, name)
	case "counter":
		CounterGetHandler(w, req, name)
	default:
		http.Error(w, "", http.StatusBadRequest)
	}
}

func GaugeGetHandler(w http.ResponseWriter, req *http.Request, name string) {
	if !st.MS.IsGaugeAllowed(name) {
		http.Error(w, "Incorrect gauge value", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	mt.Update(st.MS, &runtime.MemStats{})
	val, _ := st.MS.GetGaugeValue(name)
	if _, err := w.Write([]byte(val.String())); err != nil {
		log.Printf("error writing gauge value: %s", err)
	}
}

func CounterGetHandler(w http.ResponseWriter, req *http.Request, name string) {
	if !st.MS.IsCounterAllowed(name) {
		http.Error(w, "Incorrect counter value", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	mt.Update(st.MS, &runtime.MemStats{})
	val, _ := st.MS.GetCounterValue(name)
	if _, err := w.Write([]byte(val.String())); err != nil {
		log.Printf("error writing counter value: %s", err)
	}
}
