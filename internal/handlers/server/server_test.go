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

func TestGaugeCounterPostHandler(t *testing.T) {
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
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no-val-gauge#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no-val-gauge#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name//",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no-val-gauge#4",
			request: request{
				method: http.MethodPost,
				url:    "/update/gauge/name//fff",
			},
			want: want{
				code: http.StatusBadRequest,
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
			name: "no-name-counter#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter//3",
			},
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name: "no-val-counter#1",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no-val-counter#2",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no-val-counter#3",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name//",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "no-val-counter#4",
			request: request{
				method: http.MethodPost,
				url:    "/update/counter/name//fff",
			},
			want: want{
				code: http.StatusBadRequest,
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
		{
			name: "no-type",
			request: request{
				method: http.MethodPost,
				url:    "/update",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "incorr-type",
			request: request{
				method: http.MethodPost,
				url:    "/update/name",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "incorr-path",
			request: request{
				method: http.MethodPost,
				url:    "/name/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	r := chi.NewRouter()
	r.Route("/update", func(r chi.Router) {
		r.Route("/gauge", func(r chi.Router) {
			r.Post("/", NoNameHandler)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/*", NoValueHandler)
				r.Post("/{value}", GaugePostHandler)
			})
		})
		r.Route("/counter", func(r chi.Router) {
			r.Post("/", NoNameHandler)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/*", NoValueHandler)
				r.Post("/{value}", CounterPostHandler)
			})
		})
		r.Post("/*", GeneralCaseHandler)
	})
	r.Handle("/*", http.HandlerFunc(GeneralCaseHandler))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			//			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
			//			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestGeneralCaseHandler(t *testing.T) {
	type want struct {
		code int
		//		contentType string
	}

	type request struct {
		method string
		url    string
		//		contentType string
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
		r.Route("/gauge", func(r chi.Router) {
			r.Get("/", NoNameHandler)
			r.Get("/{name}", GaugeGetHandler)
		})
		r.Route("/counter", func(r chi.Router) {
			r.Get("/", NoNameHandler)
			r.Get("/{name}", CounterGetHandler)
		})
	})
	r.Handle("/", http.HandlerFunc(GeneralCaseHandler))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			//			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
			//			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}

func TestAllValueHandler(t *testing.T) {
	type want struct {
		code int
		//		contentType string
	}

	type request struct {
		method string
		url    string
		//		contentType string
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
		{
			name: "correct#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/meow",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	r := chi.NewRouter()
	r.Route("/value", func(r chi.Router) {
		r.Get("/*", AllValueHandler)
		r.Route("/gauge", func(r chi.Router) {
			r.Get("/*", NoNameHandler)
			r.Get("/{name}", GaugeGetHandler)
		})
		r.Route("/counter", func(r chi.Router) {
			r.Get("/*", NoNameHandler)
			r.Get("/{name}", CounterGetHandler)
		})
	})
	r.Handle("/", http.HandlerFunc(GeneralCaseHandler))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)
			//			request.Header.Set("Content-Type", test.request.contentType)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to lcose response body: %s", err)
				}
			}()
			//			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
