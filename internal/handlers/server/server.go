package server

import (
	"net/http"
	"strconv"
	"strings"

	st "github.com/a-palonskaa/metrics-server/internal/storage"
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

func MakeHandler(fn func(http.ResponseWriter, *http.Request, string, string)) http.HandlerFunc {
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
