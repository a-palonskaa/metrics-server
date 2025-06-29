package main

import (
	"net/http"
	"strconv"
	"strings"
)

type Gauge float64
type Counter int64

type MemStorage struct {
	Gauge   map[string]Gauge
	Counter map[string]Counter
}

type MetricsStorage interface {
	AddGauge(string, Gauge)
	AddCounter(string, Counter)
}

func (ms *MemStorage) AddGauge(name string, val Gauge) {
	ms.Gauge[name] = val
}

func (ms *MemStorage) AddCounter(name string, val Counter) {
	ms.Counter[name] += val
}

func gaugeHandler(w http.ResponseWriter, req *http.Request, name string, val string) {
	gaugeValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
		return
	}
	MS.AddGauge(name, Gauge(gaugeValue))
	w.WriteHeader(http.StatusOK)
}

func counterHandler(w http.ResponseWriter, req *http.Request, name string, val string) {
	counterValue, err := strconv.Atoi(val)
	if err != nil {
		http.Error(w, "Incorrect couner value", http.StatusBadRequest)
		return
	}
	MS.AddCounter(name, Counter(counterValue))
	w.WriteHeader(http.StatusOK)
}

func generalCaseHandler(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
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
			http.Error(w, "Metric value is required", http.StatusBadRequest)
			return
		}
		fn(w, r, segments[2], segments[3])
	}
}

var MS = MemStorage{
	Gauge:   make(map[string]Gauge),
	Counter: make(map[string]Counter),
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/gauge/`, makeHandler(gaugeHandler))
	mux.HandleFunc(`/update/counter/`, makeHandler(counterHandler))
	mux.HandleFunc(`/`, generalCaseHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
