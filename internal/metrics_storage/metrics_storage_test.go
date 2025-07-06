package metrics_storage

import (
	"testing"
)

//----------------------Test-MemStorage-Methods----------------------

func TestMemStorage_AddGauge(t *testing.T) {
	type fields struct {
		Gauge   map[string]Gauge
		Counter map[string]Counter
	}
	type args struct {
		name string
		val  Gauge
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{},
			},
			args: args{
				name: "name",
				val:  123.1,
			},
		},
		{
			name: "empty-name-to-empy-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{},
			},
			args: args{
				name: "",
				val:  123.1,
			},
		},
		{
			name: "existed-name-to-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]Gauge{"name": 12.0977},
				Counter: map[string]Counter{},
			},
			args: args{
				name: "name",
				val:  123.1111,
			},
		},
		{
			name: "existed-name-to-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{"name": 12.09},
				Counter: map[string]Counter{"counter": 1},
			},
			args: args{
				name: "name",
				val:  123.9,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]Gauge{"noname": 12.123},
				Counter: map[string]Counter{},
			},
			args: args{
				name: "name",
				val:  123.2,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{"noname": 12.123},
				Counter: map[string]Counter{"counter": 1},
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
				GaugeMetrics:   tt.fields.Gauge,
				CounterMetrics: tt.fields.Counter,
			}
			ms.AddGauge(tt.args.name, tt.args.val)
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	type fields struct {
		Gauge   map[string]Gauge
		Counter map[string]Counter
	}
	type args struct {
		name string
		val  Counter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "empty-name-to-empy-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{},
			},
			args: args{
				name: "",
				val:  123,
			},
		},
		{
			name: "existed-name-to-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{"counter": 12},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "existed-name-to-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{"name": 12},
				Counter: map[string]Counter{"counter": 1},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]Gauge{},
				Counter: map[string]Counter{"nocounter": 12},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage",
			fields: fields{
				Gauge:   map[string]Gauge{"noname": 12},
				Counter: map[string]Counter{"counter": 1},
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
				GaugeMetrics:   tt.fields.Gauge,
				CounterMetrics: tt.fields.Counter,
			}
			ms.AddCounter(tt.args.name, tt.args.val)
		})
	}
}
