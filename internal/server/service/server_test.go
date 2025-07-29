package server_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	server "github.com/a-palonskaa/metrics-server/internal/server/service"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
	logger "github.com/a-palonskaa/metrics-server/pkg/logger"
)

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

	msUsecase := usecase.NewMemStorage(memstorage.New())
	pingUsecase := usecase.NewPing(database.NewConn(""))

	serverHandler := server.New(server.Params{
		MsUsecase:   msUsecase,
		PingUsecase: pingUsecase,
	})

	_ = serverHandler.Router(r)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.request.method, test.request.url, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, test.code, res.StatusCode)

			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Printf("failed to close response body: %s", err)
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
			name: "working-incorr-name#1",
			request: request{
				method: http.MethodGet,
				url:    "/value/counter/name2",
			},
			code: http.StatusNotFound,
		},
	}

	r := chi.NewRouter()
	msUsecase := usecase.NewMemStorage(memstorage.New())
	pingUsecase := usecase.NewPing(database.NewConn(""))

	serverHandler := server.New(server.Params{
		MsUsecase:   msUsecase,
		PingUsecase: pingUsecase,
	})

	_ = serverHandler.Router(r)

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

	msUsecase := usecase.NewMemStorage(memstorage.New())
	pingUsecase := usecase.NewPing(database.NewConn(""))

	serverHandler := server.New(server.Params{
		MsUsecase:   msUsecase,
		PingUsecase: pingUsecase,
	})

	_ = serverHandler.Router(r)

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

func BenchmarkWithLogging(b *testing.B) {
	logger.InitLogger("../../../logs/info.log")
	handler := server.WithLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkWithCompression(b *testing.B) {
	handler := server.WithCompression(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rr := httptest.NewRecorder()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkGetAllStoredMetrics(b *testing.B) {
	msUsecase := usecase.NewMemStorage(memstorage.New())
	pingUsecase := usecase.NewPing(database.NewConn(""))

	handler := server.New(server.Params{
		MsUsecase:   msUsecase,
		PingUsecase: pingUsecase,
	})

	_ = msUsecase.UpdateMetrics(context.TODO(), []metrics.Metric{
		{ID: "heap", MType: "gauge", Value: float64(123.45)},
		{ID: "requests", MType: "counter", Delta: int64(99)},
	})

	req := httptest.NewRequest("GET", "/value/", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.GetAllStoredMetrics(rr, req)
	}
}
