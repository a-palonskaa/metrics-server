package factory

import (
	"github.com/rs/zerolog/log"

	repo "github.com/a-palonskaa/metrics-server/internal/repository"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database/storage"
	file "github.com/a-palonskaa/metrics-server/internal/repository/file"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
)

type NewParams struct {
	DatabaseAddr  string
	FilePath      string
	StoreInterval int
	Restore       bool
}

func New(p NewParams) repo.MemStorage {
	var metricsStorage repo.MemStorage
	if p.DatabaseAddr != "" {
		var ok bool
		if metricsStorage, ok = database.New(p.DatabaseAddr); !ok {
			log.Error().Msg("failed to create a db")
			return nil
		}
	} else {
		metricsStorage = memstorage.New()
	}
	return file.New(p.FilePath, metricsStorage, p.StoreInterval, p.Restore)
}
