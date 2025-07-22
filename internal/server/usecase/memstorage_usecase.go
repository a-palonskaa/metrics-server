package usecase

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"time"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	repo "github.com/a-palonskaa/metrics-server/internal/repository"
)

type MemStorageUsecaseInterface interface {
	AddMetricsToStorage(ctx context.Context, metric *metrics.Metrics) int
	GetValueFromStorage(ctx context.Context, metric *metrics.Metric) (string, int)
	UpdateValueInStorage(ctx context.Context, val *fmt.Stringer, mType string, name string) (string, int)
	WriteMetricsStorage(ctx context.Context) error
	GetAllMetrics(ctx context.Context) (map[string]float64, map[string]int64)
	Close() error
}

type MemStorageUsecase struct {
	storage       repo.MemStorage
	backup        repo.BackupStorage
	storeInterval int
}

func NewMemStorageUsecase(storage repo.MemStorage, backup repo.BackupStorage, storeInterval int, restore bool) MemStorageUsecase {
	ms := MemStorageUsecase{
		storage:       storage,
		backup:        backup,
		storeInterval: storeInterval,
	}

	if storeInterval > 0 {
		ms.StartBackupRoutine()
	}

	if restore {
		data, err := ms.LoadData(context.Background())
		if err != nil {
			log.Error().Err(err).Msg("failed to restore data")
		}
		ms.AddMetricsToStorage(context.Background(), &data)
	}
	return ms
}

func (ms *MemStorageUsecase) GetAllMetrics(ctx context.Context) (map[string]float64, map[string]int64) {
	gauges := make(map[string]float64, 0)
	counters := make(map[string]int64, 0)

	mt := ms.storage.List(ctx)
	for _, metric := range mt {
		switch metric.MType {
		case metrics.GaugeName:
			gauges[metric.ID] = *metric.Value
		case metrics.CounterName:
			counters[metric.ID] = *metric.Delta
		default:
			log.Error().Msgf("unallowed type %s", metric.MType)
		}
	}
	return gauges, counters
}

func (ms MemStorageUsecase) Close() error {
	return errors.Join(ms.storage.Close(), ms.backup.Close())
}

func (ms MemStorageUsecase) AddMetricsToStorage(ctx context.Context, mt *metrics.Metrics) int {
	for _, metric := range *mt {
		switch metric.MType {
		case metrics.GaugeName:
			ms.storage.Add(ctx, metrics.GaugeName, metric.ID, metrics.Gauge(*metric.Value))
		case metrics.CounterName:
			ms.storage.Add(ctx, metrics.CounterName, metric.ID, metrics.Counter(*metric.Delta))
		default:
			return http.StatusBadRequest
		}
	}
	return http.StatusOK
}

func (ms MemStorageUsecase) GetValueFromStorage(ctx context.Context, metric *metrics.Metric) (string, int) {
	switch metric.MType {
	case metrics.GaugeName:
		val, ok := ms.storage.Get(ctx, metrics.GaugeName, metric.ID)
		if !ok {
			return metrics.GaugeName + "name is not allowed:" + metric.ID, http.StatusNotFound
		}
		gVal, _ := val.(metrics.Gauge)
		gFloatVal := float64(gVal)
		metric.Value = &gFloatVal
	case metrics.CounterName:
		val, ok := ms.storage.Get(ctx, metrics.CounterName, metric.ID)
		if !ok {
			return metrics.CounterName + "name is not allowed:" + metric.ID, http.StatusNotFound
		}
		cVal, _ := val.(metrics.Counter)
		cIntVal := int64(cVal)
		metric.Delta = &cIntVal
	default:
		return "unknown type:" + metric.MType, http.StatusBadRequest
	}
	return "", http.StatusOK
}

func (ms MemStorageUsecase) UpdateValueInStorage(ctx context.Context, val *fmt.Stringer, mType string, name string) (string, int) {
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

	var ok bool
	*val, ok = ms.storage.Get(ctx, mType, name)
	if !ok {
		return "", http.StatusNotFound
	}
	return "", http.StatusOK
}

func (ms MemStorageUsecase) LoadData(ctx context.Context) (metrics.Metrics, error) {
	data, err := ms.backup.Load(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error loading data")
		return metrics.Metrics{}, err
	}

	var mt metrics.Metrics
	if err = mt.UnmarshalJSON(data); err != nil {
		log.Error().Err(err).Msg("error decoding body from json")
		return metrics.Metrics{}, err
	}
	return mt, nil
}

func (ms MemStorageUsecase) StartBackupRoutine() {
	go func() {
		ticker := time.NewTicker(time.Duration(ms.storeInterval) * time.Second)
		done := make(chan struct{})

		for {
			select {
			case <-ticker.C:
				mt := metrics.Metrics(ms.storage.List(context.TODO()))
				data, err := mt.MarshalJSON()
				if err != nil {
					log.Error().Err(err).Msg("error encoding to json")
				}
				if err := ms.backup.Save(context.TODO(), data); err != nil {
					log.Error().Err(err).Msg("Periodic backup failed")
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (ms MemStorageUsecase) SavingHandler() func(http.Handler) http.Handler {
	if ms.storeInterval > 0 {
		return nil
	}

	return func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn.ServeHTTP(w, r)
			if err := ms.backupMetrics(r.Context()); err != nil {
				log.Error().Err(err).Msg("Sync backup failed")
			}
		})
	}
}

func (ms MemStorageUsecase) backupMetrics(ctx context.Context) error {
	metrics := metrics.Metrics(ms.storage.List(ctx))
	data, err := metrics.MarshalJSON()
	if err != nil {
		return err
	}
	return ms.backup.Save(ctx, data)
}
