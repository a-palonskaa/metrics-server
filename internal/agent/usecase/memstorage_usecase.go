package usecase

import (
	"context"
	"math/rand"
	"runtime"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	repo "github.com/a-palonskaa/metrics-server/internal/repository"
)

type MemStorageUsecase struct {
	storage repo.MemStorage
}

type MemStorageUsecaseInterface interface {
	ListAllMetrics(ctx context.Context) metrics.Metrics
	UpdateMetrics(ctx context.Context)
}

func NewMemStorageUsecase(storage repo.MemStorage) MemStorageUsecase {
	return MemStorageUsecase{
		storage: storage,
	}
}

//----------------------MemStorageUsecase-methods----------------------

func (ms MemStorageUsecase) ListAllMetrics(ctx context.Context) metrics.Metrics {
	return metrics.Metrics(ms.storage.List(ctx))
}

func (ms MemStorageUsecase) UpdateMetrics(ctx context.Context) {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	ms.storage.Add(ctx, "gauge", "Alloc", metrics.Gauge(memStats.Alloc))
	ms.storage.Add(ctx, "gauge", "BuckHashSys", metrics.Gauge(memStats.BuckHashSys))
	ms.storage.Add(ctx, "gauge", "Frees", metrics.Gauge(memStats.Frees))
	ms.storage.Add(ctx, "gauge", "GCCPUFraction", metrics.Gauge(memStats.GCCPUFraction))
	ms.storage.Add(ctx, "gauge", "GCSys", metrics.Gauge(memStats.GCSys))
	ms.storage.Add(ctx, "gauge", "HeapAlloc", metrics.Gauge(memStats.HeapAlloc))
	ms.storage.Add(ctx, "gauge", "HeapIdle", metrics.Gauge(memStats.HeapIdle))
	ms.storage.Add(ctx, "gauge", "HeapInuse", metrics.Gauge(memStats.HeapInuse))
	ms.storage.Add(ctx, "gauge", "HeapObjects", metrics.Gauge(memStats.HeapObjects))
	ms.storage.Add(ctx, "gauge", "HeapReleased", metrics.Gauge(memStats.HeapReleased))
	ms.storage.Add(ctx, "gauge", "HeapSys", metrics.Gauge(memStats.HeapSys))
	ms.storage.Add(ctx, "gauge", "LastGC", metrics.Gauge(memStats.LastGC))
	ms.storage.Add(ctx, "gauge", "Lookups", metrics.Gauge(memStats.Lookups))
	ms.storage.Add(ctx, "gauge", "MCacheInuse", metrics.Gauge(memStats.MCacheInuse))
	ms.storage.Add(ctx, "gauge", "MCacheSys", metrics.Gauge(memStats.MCacheSys))
	ms.storage.Add(ctx, "gauge", "MSpanInuse", metrics.Gauge(memStats.MSpanInuse))
	ms.storage.Add(ctx, "gauge", "MSpanSys", metrics.Gauge(memStats.MSpanSys))
	ms.storage.Add(ctx, "gauge", "Mallocs", metrics.Gauge(memStats.Mallocs))
	ms.storage.Add(ctx, "gauge", "NextGC", metrics.Gauge(memStats.NextGC))
	ms.storage.Add(ctx, "gauge", "NumForcedGC", metrics.Gauge(memStats.NumForcedGC))
	ms.storage.Add(ctx, "gauge", "NumGC", metrics.Gauge(memStats.NumGC))
	ms.storage.Add(ctx, "gauge", "OtherSys", metrics.Gauge(memStats.OtherSys))
	ms.storage.Add(ctx, "gauge", "PauseTotalNs", metrics.Gauge(memStats.PauseTotalNs))
	ms.storage.Add(ctx, "gauge", "StackInuse", metrics.Gauge(memStats.StackInuse))
	ms.storage.Add(ctx, "gauge", "StackSys", metrics.Gauge(memStats.StackSys))
	ms.storage.Add(ctx, "gauge", "Sys", metrics.Gauge(memStats.Sys))
	ms.storage.Add(ctx, "gauge", "TotalAlloc", metrics.Gauge(memStats.TotalAlloc))
	ms.storage.Add(ctx, "gauge", "RandomValue", metrics.Gauge(rand.Float64()))
	ms.storage.Add(ctx, "counter", "PollCount", metrics.Counter(1))
}
