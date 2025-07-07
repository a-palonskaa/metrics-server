package server

import (
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(statusCode int)
}

type responseData struct {
	status int
	size   int
}

type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size = size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

func WithLogging(fn func(w http.ResponseWriter, req *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		responseData := &responseData{
			status: 0,
			size:   0,
		}

		responseWriter := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		fn(&responseWriter, req)

		log.Info().Str("uri", req.RequestURI).Str("method", req.Method).Msg("request")
		log.Info().Int("status", responseData.status).Int("size", responseData.size).Msg("response")
	}
}

func PostHandler(w http.ResponseWriter, req *http.Request) {
	mType := chi.URLParam(req, "mType")
	name := chi.URLParam(req, "name")
	val := chi.URLParam(req, "value")

	if !memstorage.IsTypeAllowed(mType) {
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

	switch mType {
	case memstorage.GaugeName:
		gaugeValue, err := strconv.ParseFloat(val, 64)
		if err != nil {
			http.Error(w, "Incorrect gauge value", http.StatusBadRequest)
			return
		}
		memstorage.MS.AddGauge(name, memstorage.Gauge(gaugeValue))
	case memstorage.CounterName:
		counterValue, err := strconv.Atoi(val)
		if err != nil {
			http.Error(w, "Incorrect couner value", http.StatusBadRequest)
			return
		}
		memstorage.MS.AddCounter(name, memstorage.Counter(counterValue))
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

	metrics.Update(memstorage.MS, &runtime.MemStats{})

	const tpl = `
	<html>
	<body>
	    <h1>MetricsStorage</h1>
	    <h2>Gauge MetricsStorage</h2>
	    <table border='1' cellpadding='5' cellspacing='0'>
	        <tr><th>Name</th><th>Value</th></tr>
	        {{range $name := .AllowedGaugeNames}}
	        <tr><td>{{ $name }}</td><td>{{index $.GaugeMetrics $name}}</td></tr>
	        {{end}}
	    </table>
	    <h2>Counter MetricsStorage</h2>
	    <table border='1' cellpadding='5' cellspacing='0'>
	        <tr><th>Name</th><th>Value</th></tr>
	        {{range $name := .AllowedCounterNames}}
	        <tr><td>{{ $name }}</td><td>{{index $.CounterMetrics $name}}</td></tr>
	        {{end}}
	    </table>
	</body>
	</html>`

	t, err := template.New("metrics").Parse(tpl)
	if err != nil {
		log.Fatal().Err(err)
	}

	err = t.Execute(w, memstorage.MS)
	if err != nil {
		log.Fatal().Err(err)
	}
}

func GetHandler(w http.ResponseWriter, req *http.Request) {
	mType := chi.URLParam(req, "mType")
	name := chi.URLParam(req, "name")

	var val fmt.Stringer
	switch mType {
	case memstorage.GaugeName:
		if !memstorage.MS.IsGaugeAllowed(name) {
			http.Error(w, "Incorrect gauge value", http.StatusNotFound)
			return
		}
		metrics.Update(memstorage.MS, &runtime.MemStats{})
		val, _ = memstorage.MS.GetGaugeValue(name)
	case memstorage.CounterName:
		if !memstorage.MS.IsCounterAllowed(name) {
			http.Error(w, "Incorrect counter value", http.StatusNotFound)
			return
		}

		metrics.Update(memstorage.MS, &runtime.MemStats{})
		val, _ = memstorage.MS.GetCounterValue(name)
	default:
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(val.String())); err != nil {
		log.Printf("error writing value: %s", err)
	}
}
