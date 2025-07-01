package metrics

import (
	"testing"

	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

//----------------------Test-MemStorage-Methods----------------------

func TestMemStorage_AddGauge(t *testing.T) {
	type fields struct {
		Gauge   map[string]st.Gauge
		Counter map[string]st.Counter
	}
	type args struct {
		name string
		val  st.Gauge
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{},
				Counter: map[string]st.Counter{},
			},
			args: args{
				name: "name",
				val:  123.1,
			},
		},
		{
			name: "empty-name-to-empy-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{},
				Counter: map[string]st.Counter{},
			},
			args: args{
				name: "",
				val:  123.1,
			},
		},
		{
			name: "existed-name-to-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]st.Gauge{"name": 12.0977},
				Counter: map[string]st.Counter{},
			},
			args: args{
				name: "name",
				val:  123.1111,
			},
		},
		{
			name: "existed-name-to-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{"name": 12.09},
				Counter: map[string]st.Counter{"counter": 1},
			},
			args: args{
				name: "name",
				val:  123.9,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]st.Gauge{"noname": 12.123},
				Counter: map[string]st.Counter{},
			},
			args: args{
				name: "name",
				val:  123.2,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{"noname": 12.123},
				Counter: map[string]st.Counter{"counter": 1},
			},
			args: args{
				name: "name",
				val:  123.2,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &st.MetricsStorage{
				GaugeMetrics:   tt.fields.Gauge,
				CounterMetrics: tt.fields.Counter,
			}
			ms.AddGauge(tt.args.name, tt.args.val)
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	type fields struct {
		Gauge   map[string]st.Gauge
		Counter map[string]st.Counter
	}
	type args struct {
		name string
		val  st.Counter
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "empty-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{},
				Counter: map[string]st.Counter{},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "empty-name-to-empy-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{},
				Counter: map[string]st.Counter{},
			},
			args: args{
				name: "",
				val:  123,
			},
		},
		{
			name: "existed-name-to-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]st.Gauge{},
				Counter: map[string]st.Counter{"counter": 12},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "existed-name-to-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{"name": 12},
				Counter: map[string]st.Counter{"counter": 1},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage-empty-counter",
			fields: fields{
				Gauge:   map[string]st.Gauge{},
				Counter: map[string]st.Counter{"nocounter": 12},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
		{
			name: "non-existed-name-to-non-empty-memStorage",
			fields: fields{
				Gauge:   map[string]st.Gauge{"noname": 12},
				Counter: map[string]st.Counter{"counter": 1},
			},
			args: args{
				name: "counter",
				val:  123,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &st.MetricsStorage{
				GaugeMetrics:   tt.fields.Gauge,
				CounterMetrics: tt.fields.Counter,
			}
			ms.AddCounter(tt.args.name, tt.args.val)
		})
	}
}
