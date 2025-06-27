package main

import (
	"net/http"
	"regexp"
	"strconv"
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
	if req.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	gaugeValue, err := strconv.ParseFloat(val, 64)
	if err != nil {
		http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
		return
	}
	MS.AddGauge(name, Gauge(gaugeValue))
	w.WriteHeader(http.StatusOK)
}

func counterHandler(w http.ResponseWriter, req *http.Request, name string, val string) {
	if req.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	counterValue, err := strconv.Atoi(val)
	if err != nil {
		http.Error(w, "Incorrect couner value", http.StatusBadRequest)
		return
	}
	MS.AddCounter(name, Counter(counterValue))
	w.WriteHeader(http.StatusOK)
}

func generalCaseHandler(w http.ResponseWriter, req *http.Request, name string, val string) {
	http.Error(w, "", http.StatusBadRequest)
}

var validPath = regexp.MustCompile(`^/update/(gauge|counter)(?:/([^/]*)(?:/(.*))?)?$`)

func makeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		metricName := m[2]
		metricValue := m[3]

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		if metricValue == "" {
			http.Error(w, "Metric value is required", http.StatusBadRequest)
			return
		}

		fn(w, r, metricName, metricValue)
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
	mux.HandleFunc(`/`, makeHandler(generalCaseHandler))

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
