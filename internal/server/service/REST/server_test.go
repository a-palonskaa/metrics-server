package serverrest_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	server "github.com/a-palonskaa/metrics-server/internal/server/service/REST"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
	hash "github.com/a-palonskaa/metrics-server/pkg/hash"
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

func TestCheckHash(t *testing.T) {
	type request struct {
		method string
		url    string
		body   []byte
		corr   bool
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
				url:    "/",
				body:   []byte(`{"key":"value"}`),
				corr:   true,
			},
			code: http.StatusOK,
		},
		{
			name: "incorrect#1",
			request: request{
				method: http.MethodGet,
				url:    "/",
				body:   []byte(`{"key":"value"}`),
				corr:   false,
			},
			code: http.StatusBadRequest,
		},
	}

	r := chi.NewRouter()
	r.Use(server.CheckHash("aliffka"))
	r.Get("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("meow"))
	}))

	key := "aliffka"
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(test.request.method, test.request.url, bytes.NewReader(test.request.body))
			if test.request.corr {
				signature, err := hash.Calculate([]byte(key), test.request.body)
				require.NoError(t, err)
				req.Header.Set("HashSHA256", hex.EncodeToString(signature))
			} else {
				req.Header.Set("HashSHA256", hex.EncodeToString([]byte("just some incorr hash")))
			}

			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			res := w.Result()
			defer func() {
				if err := res.Body.Close(); err != nil {
					log.Info().Err(err).Msg("failed to close response body")
				}
			}()

			assert.Equal(t, test.code, res.StatusCode)

			if test.request.corr {
				respHash := res.Header.Get("HashSHA256")
				require.NotEmpty(t, respHash, "response should contain HashSHA256 header")

				respBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)

				expectedHash, err := hash.Calculate([]byte(key), respBody)
				require.NoError(t, err)

				require.Equal(t, hex.EncodeToString(expectedHash), respHash)
			}
		})
	}
}

func TestWithLoggingMiddleware(t *testing.T) {
	logFile := "test_log.log"
	defer func() {
		if err := os.Remove(logFile); err != nil {
			log.Info().Err(err).Msg("failed to remove logFile")
		}
	}()

	logger.InitLogger(logFile)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	loggedHandler := server.WithLogging(handler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	loggedHandler.ServeHTTP(w, req)

	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Info().Err(err).Msg("failed to close response body")
		}
	}()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	logData, err := os.ReadFile(logFile)
	require.NoError(t, err)

	logContent := string(logData)

	assert.Contains(t, logContent, `"resp status":200`)
	assert.Contains(t, logContent, `"resp size":2`)
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
