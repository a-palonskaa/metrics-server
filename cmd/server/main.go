package main

import (
	"github.com/rs/zerolog/log"

	logger "github.com/a-palonskaa/metrics-server/internal/logger"
)

func main() {
	logger.InitLogger("info.log")

	if err := cmd.Execute(); err != nil {
    log.Fatal().Err(err)
	}
}
