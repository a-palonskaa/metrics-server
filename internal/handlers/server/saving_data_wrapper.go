package server

import (
	"net/http"
	"os"

	"github.com/rs/zerolog/log"

	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func MakeSavingHandler(ostream *os.File) func(fn http.Handler) http.Handler {
	return func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn.ServeHTTP(w, r)
			if err := memstorage.WriteMetricsStorage(ostream); err != nil {
				log.Error().Err(err)
			}
		})
	}
}
