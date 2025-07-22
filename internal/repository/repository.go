package repository

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
)

type MemStorage interface {
	Add(ctx context.Context, mType, name string, val fmt.Stringer)
	Get(ctx context.Context, mType, name string) (fmt.Stringer, bool)
	List(ctx context.Context) []metrics.Metric
	Close() error
}

type BackupStorage interface {
	Save(ctx context.Context, data []byte) error
	Load(ctx context.Context) ([]byte, error)
	Close() error
}

func NewMemStorage(databaseAddr string) MemStorage {
	var memStorage MemStorage
	if databaseAddr != "" {
		var ok bool
		if memStorage, ok = database.NewMyDB(databaseAddr); !ok {
			log.Error().Msg("failed to create a db")
			return nil
		}
	} else {
		memStorage = memstorage.NewMetricsStorage()
	}
	return memStorage
}
