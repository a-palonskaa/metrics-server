package metricsstorage

import (
	"context"
	"testing"

	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
)

//----------------------Test-MemStorage-Methods----------------------

func TestMemStorage_AddGauge(t *testing.T) {
	type fields struct {
		Gauge   map[string]metrics.Gauge
		Counter map[string]metrics.Counter
	}
	type args struct {
		name string
		val  metrics.Gauge
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{},
				Counter: map[string]metrics.Counter{},
			},
			args: args{
				name: "name",
				val:  123.1,
			},
		},
		{
			name: "empty-name-to-empy-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{},
				Counter: map[string]metrics.Counter{},
			},
			args: args{
				name: "",
				val:  123.1,
			},
		},
		{
			name: "existed-name-to-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{"name": 12.0977},
				Counter: map[string]metrics.Counter{},
			},
			args: args{
				name: "name",
				val:  123.1111,
			},
		},
		{
			name: "existed-name-to-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{"name": 12.09},
				Counter: map[string]metrics.Counter{"counter": 1},
			},
			args: args{
				name: "name",
				val:  123.9,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{"noname": 12.123},
				Counter: map[string]metrics.Counter{},
			},
			args: args{
				name: "name",
				val:  123.2,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{"noname": 12.123},
				Counter: map[string]metrics.Counter{"counter": 1},
			},
			args: args{
				name: "name",
				val:  123.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MetricsStorage{
				GaugeMetrics:        tt.fields.Gauge,
				CounterMetrics:      tt.fields.Counter,
				AllowedGaugeNames:   make(map[string]bool),
				AllowedCounterNames: make(map[string]bool),
			}
			ms.AddGauge(context.TODO(), tt.args.name, tt.args.val)
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	type fields struct {
		Gauge   map[string]metrics.Gauge
		Counter map[string]metrics.Counter
	}
	type args struct {
		name string
		val  metrics.Counter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{},
				Counter: map[string]metrics.Counter{},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "empty-name-to-empy-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{},
				Counter: map[string]metrics.Counter{},
			},
			args: args{
				name: "",
				val:  123,
			},
		},
		{
			name: "existed-name-to-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{},
				Counter: map[string]metrics.Counter{"counter": 12},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "existed-name-to-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{"name": 12},
				Counter: map[string]metrics.Counter{"counter": 1},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{},
				Counter: map[string]metrics.Counter{"nocounter": 12},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage",
			fields: fields{
				Gauge:   map[string]metrics.Gauge{"noname": 12},
				Counter: map[string]metrics.Counter{"counter": 1},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MetricsStorage{
				GaugeMetrics:        tt.fields.Gauge,
				CounterMetrics:      tt.fields.Counter,
				AllowedGaugeNames:   make(map[string]bool),
				AllowedCounterNames: make(map[string]bool),
			}
			ms.AddCounter(context.TODO(), tt.args.name, tt.args.val)
		})
	}
}
