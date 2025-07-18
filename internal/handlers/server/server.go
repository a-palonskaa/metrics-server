package server

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

// ----------------------router----------------------
func RouteRequests(r *chi.Mux, db *sql.DB, ms memstorage.MemStorage, storeInterval int, ostream *os.File, key string) *chi.Mux {
	r.Use(WithCompression)
	r.Use(WithLogging)
	r.Use(CheckHash(key))

	if storeInterval == 0 {
		r.Use(MakeSavingHandler(ostream))
	} else {
		memstorage.RunSavingStorageRoutine(ostream, storeInterval)
	}

	r.Route("/", func(r chi.Router) {
		r.Get("/", RootGetHandler)
		r.Route("/", func(r chi.Router) {
			r.Get("/ping", PingHandler(db))
			r.Post("/value/", PostJSONValueHandler(ms))
			r.Get("/value/", AllValueHandler(ms))
			r.Get("/value/{mType}/{name}", GetHandler(ms))
			r.Post("/update/", PostJSONUpdateHandler(ms))
			r.Post("/updates/", PostJSONUpdatesHandler(ms))
			r.Post("/update/{mType}/{name}/{value}", PostHandler(ms))
		})
	})
	return r
}

// ----------------------db-connection----------------------
func PingHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

//----------------------post-request-handlers----------------------

func PostHandler(ms memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		name := chi.URLParam(req, "name")
		val := chi.URLParam(req, "value")

		if message, status := validateParametrs(mType, name, val); status != http.StatusOK {
			http.Error(w, message, status)
		}

		if message, err := addValueToStorage(req.Context(), ms, mType, name, val); err != http.StatusOK {
			http.Error(w, message, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func PostJSONValueHandler(ms memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if req.ContentLength == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var metric metrics.Metrics
		var buf bytes.Buffer
		_, err := buf.ReadFrom(req.Body)
		if err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = metric.UnmarshalJSON(buf.Bytes()); err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ms.Update(req.Context(), &runtime.MemStats{})
		if message, status := getMetricValue(req.Context(), ms, &metric); status != http.StatusOK {
			log.Error().Msg(message)
			w.WriteHeader(status)
			return
		}

		resp, err := metric.MarshalJSON()
		if err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(resp); err != nil {
			log.Error().Err(err).Msg("error writing response")
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func PostJSONUpdateHandler(ms memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			log.Error().Msg("JSON format is required")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if req.ContentLength == 0 {
			log.Error().Msg("Empty body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var metric metrics.Metrics
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error().Err(err).Msg("Error Reading body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = metric.UnmarshalJSON(body); err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if ok := addMetricToStorage(req.Context(), ms, &metric); !ok {
			log.Error().Msgf("unknown type: %s", metric.MType)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp, err := metric.MarshalJSON()
		if err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(resp); err != nil {
			log.Error().Err(err).Msg("error writing response")
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

func PostJSONUpdatesHandler(ms memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		log.Info().Msg("updates request")
		contentType := req.Header.Get("Content-Type")
		if contentType != "application/json" {
			log.Error().Msg("JSON format is required")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if req.ContentLength == 0 {
			log.Error().Msg("Empty body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var metrics metrics.MetricsS
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error().Err(err).Msg("Error Reading body")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = metrics.UnmarshalJSON(body); err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if status := ms.AddMetricsToStorage(req.Context(), &metrics); status != http.StatusOK {
			w.WriteHeader(status)
			return
		}

		resp, err := metrics.MarshalJSON()
		if err != nil {
			log.Error().Err(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if _, err := w.Write(resp); err != nil {
			log.Error().Err(err).Msg("error writing response")
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
	}
}

//----------------------get-request-handlers----------------------

func GetHandler(ms memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		mType := chi.URLParam(req, "mType")
		name := chi.URLParam(req, "name")

		var val fmt.Stringer
		if message, err := updateValueInStorage(req.Context(), ms, &val, mType, name); err != http.StatusOK {
			http.Error(w, message, err)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write([]byte(val.String())); err != nil {
			log.Error().Msgf("error writing value: %s", err)
		}
	}
}
func AllValueHandler(ms memstorage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		ms.Update(req.Context(), &runtime.MemStats{})
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

		err = t.Execute(w, ms) //ХУЙНЯ -
		if err != nil {
			log.Fatal().Err(err)
		}
	}
}

func RootGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html")
}

//----------------------minor-funcs----------------------

func validateParametrs(mType string, name string, val string) (string, int) {
	if !memstorage.IsTypeAllowed(mType) {
		return "not allowed type", http.StatusBadRequest
	}

	if name == "" {
		return "empty name", http.StatusNotFound
	}

	if val == "" {
		return "empty val", http.StatusBadRequest
	}
	return "", http.StatusOK
}

func addValueToStorage(ctx context.Context, ms memstorage.MemStorage, mType string, name string, val string) (string, int) {
	switch mType {
	case metrics.GaugeName:
		gaugeValue, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return "Incorrect gauge value", http.StatusBadRequest
		}
		ms.AddGauge(ctx, name, metrics.Gauge(gaugeValue))
	case metrics.CounterName:
		counterValue, err := strconv.Atoi(val)
		if err != nil {
			return "Incorrect couner value", http.StatusBadRequest
		}
		ms.AddCounter(ctx, name, metrics.Counter(counterValue))
	}
	return "", http.StatusOK
}

func addMetricToStorage(ctx context.Context, ms memstorage.MemStorage, metric *metrics.Metrics) bool {
	switch metric.MType {
	case "gauge":
		ms.AddGauge(ctx, metric.ID, metrics.Gauge(*metric.Value))
	case "counter":
		ms.AddCounter(ctx, metric.ID, metrics.Counter(*metric.Delta))
	default:
		return false
	}
	return true
}

func updateValueInStorage(ctx context.Context, ms memstorage.MemStorage, val *fmt.Stringer, mType string, name string) (string, int) {
	switch mType {
	case metrics.GaugeName:
		log.Info().Msgf("check %s on gauge allowed", name)
		if !ms.IsGaugeAllowed(ctx, name) {
			return "Incorrect gauge value", http.StatusNotFound
		}
		ms.Update(ctx, &runtime.MemStats{})
		*val, _ = ms.GetGaugeValue(ctx, name)
	case metrics.CounterName:
		if !ms.IsCounterAllowed(ctx, name) {
			return "Incorrect counter value", http.StatusNotFound
		}

		ms.Update(ctx, &runtime.MemStats{})
		*val, _ = ms.GetCounterValue(ctx, name)
	default:
		return "not allowed type", http.StatusBadRequest
	}
	return "", http.StatusOK
}

func getMetricValue(ctx context.Context, ms memstorage.MemStorage, metric *metrics.Metrics) (string, int) {
	switch metric.MType {
	case "gauge":
		val, ok := ms.GetGaugeValue(ctx, metric.ID)
		if !ok {
			return "gauge name is not allowed:" + metric.ID, http.StatusNotFound
		}
		gVal := float64(val)
		metric.Value = &gVal
	case "counter":
		val, ok := ms.GetCounterValue(ctx, metric.ID)
		if !ok {
			return "counter name is not allowed:" + metric.ID, http.StatusNotFound
		}
		cVal := int64(val)
		metric.Delta = &cVal
	default:
		return "unknown type:" + metric.MType, http.StatusBadRequest
	}
	return "", http.StatusOK
}
