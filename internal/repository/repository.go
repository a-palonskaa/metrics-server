package factory

import (
	"github.com/rs/zerolog/log"

	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	file "github.com/a-palonskaa/metrics-server/internal/repository/file"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

type NewParams struct {
	DatabaseAddr  string
	FilePath      string
	StoreInterval int
	Restore       bool
}

func New(p NewParams) usecase.MetricsRepository {
	var metricsStorage usecase.MetricsRepository
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
