package server

import (
	"github.com/rs/zerolog/log"
	"net/http"

	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func MakeSavingHandler(p *memstorage.Producer) func(fn http.Handler) http.Handler {
	return func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn.ServeHTTP(w, r)
			if err := p.WriteStorage(); err != nil {
				log.Error().Err(err)
			}
		})
	}
}
