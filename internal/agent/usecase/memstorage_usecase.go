package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
)

type MemStorage struct {
	storage MetricsRepository
}

func NewMemStorageUsecase(storage MetricsRepository) MemStorage {
	return MemStorage{
		storage: storage,
	}
}

func (ms MemStorage) ListAllMetrics(ctx context.Context) metrics.Metrics {
	return metrics.Metrics(ms.storage.List(ctx))
}

func (ms MemStorage) UpdateMetrics(ctx context.Context) error {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)

	metricsToStore := []metrics.Metric{
		{ID: "Alloc", MType: "gauge", Value: float64(memStats.Alloc)},
		{ID: "BuckHashSys", MType: "gauge", Value: float64(memStats.BuckHashSys)},
		{ID: "Frees", MType: "gauge", Value: float64(memStats.Frees)},
		{ID: "GCCPUFraction", MType: "gauge", Value: float64(memStats.GCCPUFraction)},
		{ID: "GCSys", MType: "gauge", Value: float64(memStats.GCSys)},
		{ID: "HeapAlloc", MType: "gauge", Value: float64(memStats.HeapAlloc)},
		{ID: "HeapIdle", MType: "gauge", Value: float64(memStats.HeapIdle)},
		{ID: "HeapInuse", MType: "gauge", Value: float64(memStats.HeapInuse)},
		{ID: "HeapObjects", MType: "gauge", Value: float64(memStats.HeapObjects)},
		{ID: "HeapReleased", MType: "gauge", Value: float64(memStats.HeapReleased)},
		{ID: "HeapSys", MType: "gauge", Value: float64(memStats.HeapSys)},
		{ID: "LastGC", MType: "gauge", Value: float64(memStats.LastGC)},
		{ID: "Lookups", MType: "gauge", Value: float64(memStats.Lookups)},
		{ID: "MCacheInuse", MType: "gauge", Value: float64(memStats.MCacheInuse)},
		{ID: "MCacheSys", MType: "gauge", Value: float64(memStats.MCacheSys)},
		{ID: "MSpanInuse", MType: "gauge", Value: float64(memStats.MSpanInuse)},
		{ID: "MSpanSys", MType: "gauge", Value: float64(memStats.MSpanSys)},
		{ID: "Mallocs", MType: "gauge", Value: float64(memStats.Mallocs)},
		{ID: "NextGC", MType: "gauge", Value: float64(memStats.NextGC)},
		{ID: "NumForcedGC", MType: "gauge", Value: float64(memStats.NumForcedGC)},
		{ID: "NumGC", MType: "gauge", Value: float64(memStats.NumGC)},
		{ID: "OtherSys", MType: "gauge", Value: float64(memStats.OtherSys)},
		{ID: "PauseTotalNs", MType: "gauge", Value: float64(memStats.PauseTotalNs)},
		{ID: "StackInuse", MType: "gauge", Value: float64(memStats.StackInuse)},
		{ID: "StackSys", MType: "gauge", Value: float64(memStats.StackSys)},
		{ID: "Sys", MType: "gauge", Value: float64(memStats.Sys)},
		{ID: "TotalAlloc", MType: "gauge", Value: float64(memStats.TotalAlloc)},
		{ID: "RandomValue", MType: "gauge", Value: rand.Float64()},
		{ID: "PollCount", MType: "counter", Delta: 1},
	}

	if err := ms.storage.Update(ctx, metricsToStore); err != nil {
		log.Error().Err(err).Msg("failed to store metric")
		return fmt.Errorf("failed to store metric: %v", err)
	}
	return nil
}

func (ms MemStorage) UpdateSysMetrics(ctx context.Context) error {
	memStat, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to her vm stats")
		return err //ХУЙНЯ -  wrap
	}

	cpuStat, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		log.Error().Err(err).Msg("failed to calculate the percentage of cpu used")
		return err //ХУЙНЯ -  wrap
	}

	var metricsToStore []metrics.Metric

	metricsToStore = append(metricsToStore,
		metrics.Metric{ID: "TotalMemory", MType: "gauge", Value: float64(memStat.Total)},
		metrics.Metric{ID: "FreeMemory", MType: "gauge", Value: float64(memStat.Free)},
	)

	for i, percent := range cpuStat {
		id := fmt.Sprintf("CPUutilization%d", i+1)
		metricsToStore = append(metricsToStore, metrics.Metric{ID: id, MType: "gauge", Value: percent})
	}

	if err := ms.storage.Update(ctx, metricsToStore); err != nil {
		log.Error().Err(err).Msg("failed to store metric")
		return fmt.Errorf("failed to store metric: %v", err)
	}
	return nil
}
