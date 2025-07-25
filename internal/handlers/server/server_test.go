package server

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

//----------------------Test-Post-Handlers----------------------

func TestPostHandler(t *testing.T) {
	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name    string
		request request
		code    int
	}{
		{
			name: "no-name-gauge#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-gauge#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-gauge#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge//",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-gauge#4",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge//3",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-val-gauge#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-val-gauge#5",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/fff",
			},
			code: http.StatusBadRequest,
		},
		{
			name: "working-case-gauge#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/12.1",
			},
			code: http.StatusOK,
		},
		{
			name: "no-name-counter#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-counter#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-counter#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter//",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-val-counter#5",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name/fff",
			},
			code: http.StatusBadRequest,
		},
		{
			name: "working-case-counter#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/counter/1",
			},
			code: http.StatusOK,
		},
	}

	r := chi.NewRouter()
	r.Use(WithCompression)
	r.Use(WithLogging)

	RouteRequests(r, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
		})
	}
}

func TestGeneralCaseHandler(t *testing.T) {
	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name    string
		request request
		code    int
	}{
		{
			name: "no-name-gauge#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-gauge#2",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge/",
			},
			code: http.StatusNotFound,
		},
		{
			name: "no-name-gauge#3",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge//",
			},
			code: http.StatusNotFound,
		},
		{
			name: "non-existing-name-gauge#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge/name1",
			},
			code: http.StatusNotFound,
		},
		{
			name: "working-case-gauge#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge/Frees",
			},
			code: http.StatusOK,
		},
		{
			name: "working-case-counter#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/counter/PollCount",
			},
			code: http.StatusOK,
		},
		{
			name: "working-incorr-name#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/counter/name2",
			},
			code: http.StatusNotFound,
		},
	}

	r := chi.NewRouter()
	r.Use(WithCompression)
	r.Use(WithLogging)

	RouteRequests(r, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
		})
	}
}

func TestAllValueHandler(t *testing.T) {
	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name    string
		request request
		code    int
	}{
		{
			name: "correct#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/",
			},
			code: http.StatusOK,
		},
	}

	r := chi.NewRouter()

	r.Use(WithCompression)
	r.Use(WithLogging)

	RouteRequests(r, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
		})
	}
}
