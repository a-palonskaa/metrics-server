package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	usecase "github.com/a-palonskaa/metrics-server/internal/agent/usecase"
	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
)

func TestMemStorage_ListAllMetrics(t *testing.T) {
	type fields struct {
		storage usecase.MetricsRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   metrics.Metrics
	}{
		{
			name: "ok",
			fields: fields{
				storage: memstorage.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			want: metrics.Metrics{
				{ID: "Alloc", MType: "gauge"},
				{ID: "BuckHashSys", MType: "gauge"},
				{ID: "Frees", MType: "gauge"},
				{ID: "GCCPUFraction", MType: "gauge"},
				{ID: "GCSys", MType: "gauge"},
				{ID: "HeapAlloc", MType: "gauge"},
				{ID: "HeapIdle", MType: "gauge"},
				{ID: "HeapInuse", MType: "gauge"},
				{ID: "HeapObjects", MType: "gauge"},
				{ID: "HeapReleased", MType: "gauge"},
				{ID: "HeapSys", MType: "gauge"},
				{ID: "LastGC", MType: "gauge"},
				{ID: "Lookups", MType: "gauge"},
				{ID: "MCacheInuse", MType: "gauge"},
				{ID: "MCacheSys", MType: "gauge"},
				{ID: "MSpanInuse", MType: "gauge"},
				{ID: "MSpanSys", MType: "gauge"},
				{ID: "Mallocs", MType: "gauge"},
				{ID: "NextGC", MType: "gauge"},
				{ID: "NumForcedGC", MType: "gauge"},
				{ID: "NumGC", MType: "gauge"},
				{ID: "OtherSys", MType: "gauge"},
				{ID: "PauseTotalNs", MType: "gauge"},
				{ID: "StackInuse", MType: "gauge"},
				{ID: "StackSys", MType: "gauge"},
				{ID: "Sys", MType: "gauge"},
				{ID: "TotalAlloc", MType: "gauge"},
				{ID: "RandomValue", MType: "gauge"},
				{ID: "PollCount", MType: "counter"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := usecase.NewMemStorageUsecase(tt.fields.storage)

			err := ms.UpdateMetrics(tt.args.ctx)
			require.NoError(t, err)

			got := ms.ListAllMetrics(tt.args.ctx)

			require.Equal(t, len(got), len(tt.want))
			for _, mt := range got {
				_, err := tt.fields.storage.Get(tt.args.ctx, mt.MType, mt.ID)
				require.NoError(t, err)
			}
		})
	}
}

func TestMemStorage_UpdateMetrics(t *testing.T) {
	type fields struct {
		storage usecase.MetricsRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				storage: memstorage.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := usecase.NewMemStorageUsecase(tt.fields.storage)
			err := ms.UpdateMetrics(tt.args.ctx)
			if !tt.wantErr {
				require.NoError(t, err)
			} else if err == nil {
				t.Errorf("must be an error")
			}
		})
	}
}

func TestMemStorage_UpdateSysMetrics(t *testing.T) {
	type fields struct {
		storage usecase.MetricsRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ok",
			fields: fields{
				storage: memstorage.New(),
			},
			args: args{
				ctx: context.TODO(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := usecase.NewMemStorageUsecase(tt.fields.storage)
			err := ms.UpdateSysMetrics(tt.args.ctx)
			if !tt.wantErr {
				require.NoError(t, err)
			} else if err == nil {
				t.Errorf("must be an error")
			}
		})
	}
}

func BenchmarkUpdateMetrics(b *testing.B) {
	ctx := context.TODO()
	memStorage := usecase.NewMemStorageUsecase(memstorage.New())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = memStorage.UpdateMetrics(ctx)
	}
}

func BenchmarkSysMetrics(b *testing.B) {
	ctx := context.Background()
	memStorage := usecase.NewMemStorageUsecase(memstorage.New())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = memStorage.UpdateSysMetrics(ctx)
	}
}
