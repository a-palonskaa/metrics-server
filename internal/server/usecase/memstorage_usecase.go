package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	repo "github.com/a-palonskaa/metrics-server/internal/repository"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	errhandlers "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
)

type MemStorageUsecase struct {
	storage repo.MemStorage
	ostream *os.File
}

type MemStorageUsecaseInterface interface {
	AddMetricsToStorage(ctx context.Context, metric *metrics.Metrics) int
	GetValueFromStorage(ctx context.Context, metric *metrics.Metric) (string, int)
	UpdateValueInStorage(ctx context.Context, val *fmt.Stringer, mType string, name string) (string, int)
	WriteMetricsStorage() error
	Close() error
}

func (ms MemStorageUsecase) Close() error {
	err := ms.storage.Close()
	if err != nil {
		log.Error().Err(err).Msg("error closing storage")
	}

	if ms.ostream != os.Stdout && ms.ostream != os.Stderr {
		errOstream := ms.ostream.Close()
		if errOstream != nil {
			log.Error().Err(errOstream).Msg("error closing ostream")
			err = errors.Join(err, errOstream)
		}
	}
	return err
}

// FIXME -  extralogs
func (ms MemStorageUsecase) AddMetricsToStorage(ctx context.Context, mt *metrics.Metrics) int {
	for _, metric := range *mt {
		log.Info().Msgf("%s %s", metric.MType, metric.ID)
		switch metric.MType {
		case metrics.GaugeName:
			log.Info().Msgf("adding GAUGE metrics %s %s %v", metric.MType, metric.ID, *metric.Value)
			ms.storage.Add(ctx, metrics.GaugeName, metric.ID, metrics.Gauge(*metric.Value))
		case metrics.CounterName:
			log.Info().Msgf("adding COUNTER metrics %s %s %v", metric.MType, metric.ID, *metric.Delta)
			ms.storage.Add(ctx, metrics.CounterName, metric.ID, metrics.Counter(*metric.Delta))
		default:
			log.Info().Msgf("unknown type %s, returning BadRequest", metric.MType)
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

func (ms MemStorageUsecase) WriteMetricsStorage() error {
	if _, err := ms.ostream.Seek(0, 0); err != nil {
		log.Error().Err(err).Msgf("moving file prt to begining %v %v", ms.ostream, ms.ostream == os.Stdout)
		return err
	}

	allMetrics := metrics.Metrics(ms.storage.List(context.TODO()))
	data, err := allMetrics.MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("error encoding to json")
		return err
	}

	if _, err := ms.ostream.Write(append(data, '\n')); err != nil {
		log.Error().Err(err).Msg("error writing data to ostream")
		return err
	}
	return nil
}

func (ms *MemStorageUsecase) Init(databaseAddr string, restore bool, fileStoragePath string) {
	ms.ostream = os.Stdout
	if databaseAddr == "" {
		ms.storage = memstorage.MS
		if restore {
			if err := ms.readMetricsStorage(fileStoragePath); err != nil {
				log.Error().Err(err).Msg("error reading metrics storage")
			}
		}
	} else {
		db, err := errhandlers.RetriableErrHadler(
			func() (*sql.DB, error) { return sql.Open("pgx", databaseAddr) },
			errhandlers.CompareErrSQL,
		)
		if err != nil {
			log.Error().Err(err).Msg("failed to initialize *sql.DB and create a connection pull")
			return
		}

		err = errhandlers.RetriableErrHadlerVoid(
			func() error { return database.CreateTables(db) },
			errhandlers.CompareErrSQL)
		if err != nil {
			log.Error().Err(err).Msg("failed to create tables")
			return
		}
		ms.storage = database.CreateMyDB(db)
	}

	ostream, err := os.OpenFile(fileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		log.Error().Err(err).Msgf("error opening file %s", fileStoragePath)
		return
	}
	ms.ostream = ostream
}

func (ms *MemStorageUsecase) readMetricsStorage(filename string) error {
	istream, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error().Err(err).Msgf("error opening file %s", filename)
		return err
	}

	istreamInfo, err := istream.Stat()
	if err != nil {
		log.Error().Err(err).Msg("error getting istream info")
		return err
	}

	data := make([]byte, istreamInfo.Size())
	_, err = istream.Read(data)
	if err != nil {
		log.Error().Err(err).Msg("error reading from istream")
		return err
	}

	var allMetrics metrics.Metrics
	if err := allMetrics.UnmarshalJSON(data); err != nil {
		log.Error().Err(err).Msg("error decoding data from json")
		return err
	}
	ms.AddMetricsToStorage(context.TODO(), &allMetrics)

	if err := istream.Close(); err != nil {
		log.Error().Err(err).Msgf("error closing file %s", filename)
		return err
	}
	return nil
}
