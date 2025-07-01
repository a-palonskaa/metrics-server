package server

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	mt "github.com/a-palonskaa/metrics-server/internal/metrics"
	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func GaugePostHandler(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	if name == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if val == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	gaugeValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
		return
	}
	st.MS.AddGauge(name, st.Gauge(gaugeValue))
	w.WriteHeader(http.StatusOK)
}

func CounterPostHandler(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	if name == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}

	if val == "" {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	counterValue, err := strconv.Atoi(val)
	if err != nil {
		http.Error(w, "Incorrect couner value", http.StatusBadRequest)
		return
	}
	st.MS.AddCounter(name, st.Counter(counterValue))
	w.WriteHeader(http.StatusOK)
}

func GeneralCaseHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func NoNameHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusNotFound)
}

func NoValueHandler(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	if name == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	http.Error(w, "", http.StatusBadRequest) //DEBUG - NOVAL
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
	fmt.Fprintf(w, "<html><body><h1>MetricsStorage</h1>")
	fmt.Fprintf(w, "<h2>Gauge MetricsStorage</h2>")
	fmt.Fprintf(w, "<table border='1' cellpadding='5' cellspacing='0'>")
	fmt.Fprintf(w, "<tr><th>Name</th><th>Value</th></tr>")
	for _, key := range st.MS.AllowedGaugeNames {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td></tr>\n", key, st.MS.GaugeMetrics[key])
	}
	fmt.Fprintln(w, "</table>")

	fmt.Fprintf(w, "<h2>Counter MetricsStorage</h2>")
	fmt.Fprintf(w, "<table border='1' cellpadding='5' cellspacing='0'>")
	fmt.Fprintf(w, "<tr><th>Name</th><th>Value</th></tr>")
	for _, key := range st.MS.AllowedCounterNames {
		fmt.Fprintf(w, "<tr><td>%s</td><td>%v</td></tr>\n", key, st.MS.CounterMetrics[key])
	}
	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, "</body></html>")

}

func GaugeGetHandler(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")

	if !st.MS.IsGaugeAllowed(name) {
		http.Error(w, "Incorrect gauge value", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	mt.Update(st.MS, &runtime.MemStats{})
	fmt.Fprintf(w, "<html><body><h1>%s</h1>", name)

	val, _ := st.MS.GetGaugeValue(name)
	fmt.Fprintf(w, "<h2>value: %v</h2>\n", val)
	fmt.Fprintf(w, "</body></html>")
}

func CounterGetHandler(w http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")

	if !st.MS.IsCounterAllowed(name) {
		http.Error(w, "Incorrect counter value", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	mt.Update(st.MS, &runtime.MemStats{})
	fmt.Fprintf(w, "<html><body><h1>%s</h1>", name)

	val, _ := st.MS.GetCounterValue(name)
	fmt.Fprintf(w, "<h2>value: %v</h2>\n", val)
	fmt.Fprintf(w, "</body></html>")
}
