package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/a-palonskaa/metrics-server/internal/metrics"
	"github.com/go-resty/resty/v2"
)

func TestSendRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Error("Missing gzip content encoding")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := resty.New()

	counter := int64(1)
	gauge := float64(1.24)

	type args struct {
		client   *resty.Client
		endpoint string
		body     metrics.MetricsS
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success-case-gauge",
			args: args{
				client:   client,
				endpoint: ts.URL[7:],
				body: metrics.MetricsS{
					{
						ID:    "Frees",
						MType: "gauge",
						Value: &gauge,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success-case-counter",
			args: args{
				client:   client,
				endpoint: ts.URL[7:],
				body: metrics.MetricsS{
					{
						ID:    "Frees",
						MType: "counter",
						Delta: &counter,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "empty-body",
			args: args{
				client:   client,
				endpoint: ts.URL[7:],
				body:     metrics.MetricsS{},
			},
			wantErr: false,
		},
		{
			name: "invalid-url",
			args: args{
				client:   client,
				endpoint: "",
				body: metrics.MetricsS{
					{
						ID:    "Frees",
						MType: "counter",
						Delta: &counter,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendRequest(tt.args.client, tt.args.endpoint, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
