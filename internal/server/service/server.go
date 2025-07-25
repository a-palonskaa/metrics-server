package server

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
	MsUsecase   usecase.MemStorage
	PingUsecase usecase.Ping
}

type ServerHandler struct {
	msUsecase   usecase.MemStorage
	pingUsecase usecase.Ping
}

func New(p Params) ServerHandler {
	return ServerHandler{
		msUsecase:   p.MsUsecase,
		pingUsecase: p.PingUsecase,
	}
}

func (h ServerHandler) Router(r *chi.Mux) *chi.Mux {
	r.Use(WithCompression)
	r.Use(WithLogging)

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

func (h ServerHandler) CheckConnection(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
	defer cancel()

	if err := h.pingUsecase.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

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
	w.WriteHeader(http.StatusOK)
}

func (h ServerHandler) GetMetricByBody(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.ContentLength == 0 {
		w.WriteHeader(http.StatusBadRequest)
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

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(resp); err != nil {
		log.Error().Err(err).Msg("error writing response")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

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
	w.WriteHeader(http.StatusOK)
}

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
	w.WriteHeader(http.StatusOK)
}

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

func (h ServerHandler) GetAllStoredMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	metricsList := h.msUsecase.GetAllMetrics(r.Context())

	data := struct {
		Metrics []metrics.Metric
	}{
		Metrics: metricsList,
	}

	t := template.Must(template.New("metrics").Parse(htmlTemplate))
	if err := t.Execute(w, data); err != nil {
		log.Error().Err(err).Msg("error executing template")
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}

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
