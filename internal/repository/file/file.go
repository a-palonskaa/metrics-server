// Package file implements a MetricsRepository interface with memtrics storage and backup file storage.
package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

// FileStorage provides a metrics repository with file-based backup and restore.
type FileStorage struct {
	mu      sync.RWMutex
	file    *os.File
	storage usecase.MetricsRepository
}

// New creates a new FileStorage instance.
func New(path string, storage usecase.MetricsRepository, storeInterval int, restore bool) *FileStorage {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Error().Err(err).Msg("failed to open backup file")
		file = os.Stdout
	}

	fs := FileStorage{
		file:    file,
		storage: storage,
	}

	if storeInterval > 0 {
		fs.StartBackupRoutine(storeInterval)
	}

	if restore {
		data, err := fs.LoadData(context.TODO())
		if err != nil {
			log.Error().Err(err).Msg("failed to restore data")
		}
		for _, metric := range data {
			if err := fs.storage.Update(context.TODO(), metrics.Metrics([]metrics.Metric{metric})); err != nil {
				log.Error().Err(err).Msgf("error adding metric %v", metric)
			}
		}
	}
	return &fs
}

func (fs *FileStorage) Update(ctx context.Context, metric metrics.Metrics) error {
	if err := fs.storage.Update(ctx, metric); err != nil {
		log.Error().Err(err).Msg("failed to add metric")
		return err
	}

	if err := fs.Save(ctx); err != nil {
		log.Error().Err(err).Msg("backup failed")
		return err
	}
	return nil
}

func (fs *FileStorage) Get(ctx context.Context, mType, name string) (metrics.Metric, error) {
	metric, err := fs.storage.Get(ctx, mType, name)
	if err != nil {
		log.Error().Err(err).Msgf("failed to get metric type:%s, name:%s", mType, name)
		return metrics.Metric{}, err
	}
	return metric, nil
}

func (fs *FileStorage) List(ctx context.Context) []metrics.Metric {
	return fs.storage.List(ctx)
}

func (fs *FileStorage) Close() error {
	return errors.Join(fs.storage.Close(), fs.file.Close())
}

func (fs *FileStorage) Load(_ context.Context) ([]byte, error) {
	istreamInfo, err := fs.file.Stat()
	if err != nil {
		log.Error().Err(err).Msg("error getting istream info")
		return make([]byte, 0), err
	}

	data := make([]byte, istreamInfo.Size())
	_, err = fs.file.Read(data)
	if err != nil {
		log.Error().Err(err).Msg("error reading from istream")
		return make([]byte, 0), err
	}
	return data, nil
}

func (fs *FileStorage) LoadData(ctx context.Context) (metrics.Metrics, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := fs.Load(ctx)
	if err != nil {
		log.Error().Err(err).Msg("error loading data")
		return metrics.Metrics{}, err
	}

	var mt metrics.RawMetrics
	if err = mt.UnmarshalJSON(data); err != nil {
		log.Error().Err(err).Msg("error decoding body from json")
		return metrics.Metrics{}, err
	}
	return mt.Serialize(), nil
}

func (fs *FileStorage) StartBackupRoutine(storeInterval int) {
	go func() {
		ticker := time.NewTicker(time.Duration(storeInterval) * time.Second)
		done := make(chan struct{})

		for {
			select {
			case <-ticker.C:
				if err := fs.Save(context.TODO()); err != nil {
					log.Error().Err(err).Msg("Periodic backup failed")
					return
				}
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()
}

func (fs *FileStorage) Save(ctx context.Context) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := metrics.Metrics(fs.storage.List(ctx)).MarshalJSON()
	if err != nil {
		log.Error().Err(err).Msg("error encoding to json")
		return fmt.Errorf("encoding to json: %v", err)
	}

	if _, err := fs.file.Seek(0, 0); err != nil {
		log.Error().Err(err).Msgf("moving file prt to begining %v", fs.file)
		return fmt.Errorf("moving file prt to begining %v", err)
	}

	if _, err := fs.file.Write(data); err != nil {
		log.Error().Err(err).Msg("error writing data to ostream")
		return fmt.Errorf("error writing data: %v", err)
	}
	return nil
}
