package agent_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/require"

	agent "github.com/a-palonskaa/metrics-server/internal/agent/service/REST"
	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
)

func TestSendRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Error("Missing gzip content encoding")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	counter := int64(1)
	gauge := float64(1.24)

	type args struct {
		body metrics.Metrics
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success-case-gauge",
			args: args{
				body: metrics.Metrics{
					{
						ID:    "Frees",
						MType: "gauge",
						Value: gauge,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success-case-counter",
			args: args{
				body: metrics.Metrics{
					{
						ID:    "Frees",
						MType: "counter",
						Delta: counter,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty-body",
			args: args{
				body: metrics.Metrics{},
			},
			wantErr: false,
		},
	}

	handler := agent.NewHandler(memstorage.New(), "", client)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.SendRequest(tt.args.body)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSendRequestWithHash(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Error("Missing gzip content encoding")
		}
		if r.Header.Get("HashSHA256") == "" {
			t.Error("Missing HashSHA256 header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	counter := int64(1)
	gauge := float64(1.24)

	tests := []struct {
		name    string
		body    metrics.Metrics
		wantErr bool
	}{
		{
			name: "with-hash-gauge",
			body: metrics.Metrics{
				{
					ID:    "Alloc",
					MType: "gauge",
					Value: gauge,
				},
			},
			wantErr: false,
		},
		{
			name: "with-hash-counter",
			body: metrics.Metrics{
				{
					ID:    "PollCount",
					MType: "counter",
					Delta: counter,
				},
			},
			wantErr: false,
		},
	}

	handler := agent.NewHandler(memstorage.New(), "key", client)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.SendRequest(tt.body)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSendMetrics(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Error("Missing gzip content encoding")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New().SetBaseURL(ts.URL)

	counter := int64(1)
	gauge := float64(1.24)

	type args struct {
		body metrics.Metrics
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success-case-gauge",
			args: args{
				body: metrics.Metrics{
					{
						ID:    "Frees",
						MType: "gauge",
						Value: gauge,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success-case-counter",
			args: args{
				body: metrics.Metrics{
					{
						ID:    "Frees",
						MType: "counter",
						Delta: counter,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty-body",
			args: args{
				body: metrics.Metrics{},
			},
			wantErr: false,
		},
	}

	handler := agent.NewHandler(memstorage.New(), "", client)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.SendMetrics(context.TODO())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func BenchmarkSendRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := resty.New().SetBaseURL(server.URL)
	handler := agent.NewHandler(memstorage.New(), "aliffka", client)

	const count = 50
	metric := make([]metrics.Metric, count)
	for i := 0; i < count; i++ {
		id := strconv.Itoa(i)
		if i%2 == 0 {
			metric[i] = metrics.Metric{
				ID:    id,
				MType: metrics.GaugeName,
				Value: float64(i),
			}
		} else {
			metric[i] = metrics.Metric{
				ID:    id,
				MType: metrics.CounterName,
				Delta: int64(i),
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = handler.SendRequest(metric)
	}
}
