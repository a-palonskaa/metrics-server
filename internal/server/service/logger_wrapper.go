package server

import (
	"net/http"

	"github.com/rs/zerolog/log"
)

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

// WithLogging creates a middleware wrapper to log request method, response status and size.
func WithLogging(fn http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		responseData := responseData{}
		responseWriter := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   &responseData,
		}

		fn.ServeHTTP(&responseWriter, req)
		log.Info().Str("uri", req.RequestURI).Str("method", req.Method).Int("resp status", responseData.status).Int("resp size", responseData.size).Msg("request")
	})
}
