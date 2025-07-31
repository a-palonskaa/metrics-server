// Package server provides router for server, http-requests handlers,
// middlewares for logging, encoding, hash signature checking
package serverREST

import (
	"bytes"
	"context"
	"errors"
	"html/template"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

var metricsTemplate = template.Must(template.New("metrics").Parse(htmlTemplate))

const htmlTemplate = `
<html>
<body>
    <h1>Metrics</h1>
    <h2>Gauge Metrics</h2>
    <table border='1'>
        {{range .Metrics}}
            {{if eq .MType "gauge"}}
            <tr>
                <td>{{.ID}}</td>
                <td>{{.Value}}</td>
            </tr>
            {{end}}
        {{end}}
    </table>

    <h2>Counter Metrics</h2>
    <table border='1'>
        {{range .Metrics}}
            {{if eq .MType "counter"}}
            <tr>
                <td>{{.ID}}</td>
                <td>{{.Delta}}</td>
            </tr>
            {{end}}
        {{end}}
    </table>
</body>
</html>`

type Params struct {
	MsUsecase     usecase.MemStorage
	PingUsecase   usecase.Ping
	Key           string
	TrustedSubnet string
}

type ServerHandler struct {
	msUsecase     usecase.MemStorage
	pingUsecase   usecase.Ping
	key           string
	trustedSubnet string
}

func New(p Params) ServerHandler {
	return ServerHandler{
		msUsecase:     p.MsUsecase,
		pingUsecase:   p.PingUsecase,
		key:           p.Key,
		trustedSubnet: p.TrustedSubnet,
	}
}

func (h ServerHandler) Router(r *chi.Mux) *chi.Mux {
	r.Use(WithCompression)
	r.Use(WithLogging)
	r.Use(ValidateIP(h.trustedSubnet))
	r.Use(CheckHash(h.key))

	r.Mount("/debug/pprof", http.DefaultServeMux)

	r.Route("/", func(r chi.Router) {
		r.Get("/", h.Root)
		r.Route("/", func(r chi.Router) {
			r.Get("/ping", h.CheckConnection)

			r.Post("/value/", h.GetMetricByBody)
			r.Get("/value/", h.GetAllStoredMetrics)
			r.Get("/value/{mType}/{name}", h.GetMetricValueByURI)

			r.Post("/update/", h.UpdateMetric)
			r.Post("/updates/", h.UpdateMetrics)
			r.Post("/update/{mType}/{name}/{value}", h.UpdateMetricByURL)
		})
	})
	return r
}

// CheckConnection check the connection to the database.
//
// @Summary Connection check
// @Description Returns 200 OK if the database is reachable, 500 if not
// @Tags Connection
// @Success 200 {string} string "OK"
// @Failure 500 {string} string "Database unreachable"
// @Router /ping [get]
func (h ServerHandler) CheckConnection(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.pingUsecase.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// UpdateMetricByURL updates a metric by name, type, and value.
//
// @Summary Update metric by URL
// @Description Updates a metric by type, name, and value provided in the URL
// @Tags Metrics
// @Accept text/plain
// @Produce text/plain
// @Success 200 {string} string "Metric updated"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /update/{mType}/{name}/{value} [post]
func (h ServerHandler) UpdateMetricByURL(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	name := chi.URLParam(r, "name")
	val := chi.URLParam(r, "value")

	if message, status := validateParametrs(mType, name, val); status != http.StatusOK {
		http.Error(w, message, status)
		log.Error().Msg(message)
		return
	}

	metric, err := metrics.NewMetric(name, mType, val)
	if err != nil {
		http.Error(w, "%v", http.StatusBadRequest)
		log.Error().Err(err).Msg("failed to create metric")
		return
	}

	if err := h.msUsecase.UpdateMetrics(r.Context(), metrics.Metrics{metric}); err != nil {
		http.Error(w, "%v", http.StatusBadRequest)
		log.Error().Err(err).Msg("failed to add metrics to storage")
		return
	}
}

// GetMetricByBody returns the value of a metric from the JSON request body.
//
// @Summary Get a metric by request body
// @Description Returns a metric from the JSON body by type and name
// @Tags Metrics
// @Accept application/json
// @Produce application/json
// @Param metric body metrics.RawMetric true "Metric data"
// @Success 200 {object} metrics.RawMetric "Found metric"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /value/ [post]
func (h ServerHandler) GetMetricByBody(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.ContentLength == 0 {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Msg("zero len body")
		return
	}

	var mtr metrics.RawMetric
	var buf bytes.Buffer
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("error reading request body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = mtr.UnmarshalJSON(buf.Bytes()); err != nil {
		log.Error().Err(err).Msg("error decoding body from json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric := mtr.Serialize()
	if err := h.msUsecase.GetMetric(r.Context(), &metric); err != nil {
		switch {
		case errors.Is(err, metrics.ErrUnallowedMetric):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, metrics.ErrIncorrectMetricType):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		log.Error().Err(err).Msg("error getting val")
		return
	}

	mtr = metric.Deserialize()
	resp, err := mtr.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("error encoding body to json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Error().Err(err).Msg("error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// UpdateMetric updates a metric from a JSON request body.
//
// @Summary Update a metric
// @Description Updates a metric from a JSON request body (type, name, and value)
// @Tags Metrics
// @Accept application/json
// @Produce application/json
// @Param metric body metrics.RawMetric true "Metric to update"
// @Success 200 {object} metrics.RawMetric "Updated metric"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /update/ [post]
func (h ServerHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Error().Msg("JSON format is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if r.ContentLength == 0 {
		log.Error().Msg("Empty body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var mtr metrics.RawMetric
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error Reading body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = mtr.UnmarshalJSON(body); err != nil {
		log.Error().Err(err).Msg("error decoding body from json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric := mtr.Serialize()
	if err := h.msUsecase.UpdateMetrics(r.Context(), metrics.Metrics{metric}); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Msgf("failed to store metric")
		return
	}

	mtr = metric.Deserialize()
	resp, err := mtr.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("error encoding body to json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Error().Err(err).Msg("error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// UpdateMetrics updates list of metric from a JSON request body.
//
// @Summary Update list of metric
// @Description Accepts an array of metrics in JSON format and stores them
// @Tags Metrics
// @Accept application/json
// @Produce application/json
// @Param metrics body metrics.RawMetrics true "List of metrics to update"
// @Success 200 {object} metrics.RawMetrics "Updated metrics"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /updates/ [post]
func (h ServerHandler) UpdateMetrics(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Error().Msg("JSON format is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if r.ContentLength == 0 {
		log.Error().Msg("Empty body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var mtr metrics.RawMetrics
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("Error Reading body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = mtr.UnmarshalJSON(body); err != nil {
		log.Error().Err(err).Msg("error decoding body from json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metric := mtr.Serialize()
	if err := h.msUsecase.UpdateMetrics(r.Context(), metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error().Err(err).Msg("failed to store metrics")
		return
	}

	mtr = metric.Deserialize()
	resp, err := mtr.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("error encoding body to json")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(resp); err != nil {
		log.Error().Err(err).Msg("error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// GetMetricValueByURI returns the value of a metric by type and name.
//
// @Summary Get a metric by name and type
// @Description Returns a metric from the JSON body by type and name
// @Tags Metrics
// @Produce text/plain
// @Success 200 {string} string "Metric value"
// @Failure 404 {string} string "Metric not found"
// @Failure 500 {string} string "Internal Server Error"
// @Router /value/{mType}/{name} [get]
func (h ServerHandler) GetMetricValueByURI(w http.ResponseWriter, r *http.Request) {
	mType := chi.URLParam(r, "mType")
	name := chi.URLParam(r, "name")

	metric := metrics.Metric{
		ID:    name,
		MType: mType,
	}

	if err := h.msUsecase.GetMetric(r.Context(), &metric); err != nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		log.Error().Err(err).Msg("failed to get metric")
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	if _, err := w.Write([]byte(metric.StrValue())); err != nil {
		log.Error().Err(err).Msg("error writing value")
	}
}

// GetAllStoredMetrics returns all stored metrics in an HTML table.
//
// @Summary Returns a list of all metrics
// @Description Renders all stored gauge and counter metrics as an HTML table
// @Tags Metrics
// @Produce text/html
// @Success 200 {string} string "HTML with list of metrics"
// @Failure 500 {string} string "Template error"
// @Router /value/ [get]
func (h ServerHandler) GetAllStoredMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	metricsList := h.msUsecase.GetAllMetrics(r.Context())

	data := struct {
		Metrics []metrics.Metric
	}{
		Metrics: metricsList,
	}

	if err := metricsTemplate.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("error executing template")
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

// Root shows a root page.
//
// @Summary Root page
// @Description Returns 200 OK if request method is GET
// @Tags Root
// @Success 200 {string} string "OK"
// @Failure 400 {string} string "Method not allowed"
// @Router / [get]
func (h ServerHandler) Root(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html")
}

func validateParametrs(mType string, name string, val string) (string, int) {
	if !metrics.IsTypeAllowed(mType) {
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
