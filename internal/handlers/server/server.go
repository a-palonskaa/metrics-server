package server

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	mt "github.com/a-palonskaa/metrics-server/internal/metrics"
	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func GaugeHandler(w http.ResponseWriter, req *http.Request, name string, val string) {
	gaugeValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
		return
	}
	st.MS.AddGauge(name, st.Gauge(gaugeValue))
	w.WriteHeader(http.StatusOK)
}

func CounterHandler(w http.ResponseWriter, req *http.Request, name string, val string) {
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

func MakePostHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(segments) == 2 || segments[2] == "" {
			http.Error(w, "Invalid path format", http.StatusNotFound)
			return
		}

		if len(segments) == 3 || segments[3] == "" {
			http.Error(w, "Val s required", http.StatusBadRequest)
			return
		}
		fn(w, r, segments[2], segments[3])
	}
}

func MakeGetHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only Get value-requests are allowed!", http.StatusMethodNotAllowed)
			return
		}

		segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(segments) == 2 || segments[1] == "" {
			http.Error(w, "Metric type is required", http.StatusNotFound)
			return
		}

		if segments[2] == "" {
			http.Error(w, "Metric name is required", http.StatusBadRequest)
			return
		}
		fn(w, r, segments[2])
	}
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

func GaugeValueHandler(w http.ResponseWriter, req *http.Request, name string) {
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

func CounterValueHandler(w http.ResponseWriter, req *http.Request, name string) {
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
