package server

import (
	"net/http"

	"github.com/rs/zerolog/log"
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
