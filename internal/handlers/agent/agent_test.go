package agent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
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

	type args struct {
		client   *resty.Client
		endpoint string
		mType    string
		name     string
		val      fmt.Stringer
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
				mType:    "gauge",
				name:     "Frees",
				val:      metrics.Gauge(1.54),
			},
			wantErr: false,
		},
		{
			name: "success-case-counter",
			args: args{
				client:   client,
				endpoint: ts.URL[7:],
				mType:    "counter",
				name:     "Frees",
				val:      metrics.Counter(5),
			},
			wantErr: false,
		},
		{
			name: "invalid-type-case",
			args: args{
				client:   client,
				endpoint: ts.URL[7:],
				mType:    "invalid",
				name:     "Frees",
				val:      metrics.Gauge(1.54),
			},
			wantErr: true,
		},
		{
			name: "invalid-url-case",
			args: args{
				client:   client,
				endpoint: "",
				mType:    "counter",
				name:     "Frees",
				val:      metrics.Counter(5),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendRequest(tt.args.client, tt.args.endpoint, tt.args.mType, tt.args.name, tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
