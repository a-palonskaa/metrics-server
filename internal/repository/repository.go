// Package repository provides a constructor for initializing the metrics storage
// based on configuration parameters.
package repository

import (
	"github.com/rs/zerolog/log"

	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	file "github.com/a-palonskaa/metrics-server/internal/repository/file"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

// NewParams holds configuration parameters for initializing a metrics repository.
type NewParams struct {
	DatabaseAddr  string
	FilePath      string
	StoreInterval int
	Restore       bool
}

// New initializes and returns a MetricsRepository based on the provided parameters.
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
