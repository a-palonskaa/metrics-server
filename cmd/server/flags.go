package main

import (
	"github.com/rs/zerolog/log"

	logger "github.com/a-palonskaa/metrics-server/pkg/logger"
)

func main() {
	logger.InitLogger("logs/info_server.log")

	if err := cmd.Execute(); err != nil {
		log.Error().Err(err)
		return
	}
}
