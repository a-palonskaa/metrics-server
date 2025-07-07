package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

//----------------------Test-Post-Handlers----------------------

func TestPostHandler(t *testing.T) {
	type want struct {
		code int
	}

	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "no-name-gauge#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-gauge#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-gauge#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge//",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-gauge#4",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge//3",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-val-gauge#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-val-gauge#5",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/fff",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "working-case-gauge#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/12.1",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "no-name-counter#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-counter#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-counter#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter//",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-val-counter#5",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name/fff",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "working-case-counter#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/counter/1",
			},
			want: want{
				code: http.StatusOK,
			},
		},
	}

	r := chi.NewRouter()
	r.Route("/value", func(r chi.Router) {
		r.Get("/", AllValueHandler)
		r.Get("/{mType}/{name}", GetHandler)
	})
	r.Post("/update/{mType}/{name}/{value}", PostHandler)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
		})
	}
}

func TestGeneralCaseHandler(t *testing.T) {
	type want struct {
		code int
	}

	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "no-name-gauge#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-gauge#2",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge/",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-name-gauge#3",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge//",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "non-existing-name-gauge#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge/name1",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "working-case-gauge#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/gauge/Frees",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "working-case-counter#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/counter/PollCount",
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "working-incorr-name#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/counter/name2",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
	}

	r := chi.NewRouter()
	r.Route("/value", func(r chi.Router) {
		r.Get("/", AllValueHandler)
		r.Get("/{mType}/{name}", GetHandler)
	})
	r.Post("/update/{mType}/{name}/{value}", PostHandler)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
		})
	}
}

func TestAllValueHandler(t *testing.T) {
	type want struct {
		code int
	}

	type request struct {
		method string
		url    string
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "correct#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/",
			},
			want: want{
				code: http.StatusOK,
			},
		},
	}

	r := chi.NewRouter()
	r.Route("/value", func(r chi.Router) {
		r.Get("/", AllValueHandler)
		r.Get("/{mType}/{name}", GetHandler)
	})
	r.Post("/update/{mType}/{name}/{value}", PostHandler)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
		})
	}
}
