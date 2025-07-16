package main

import (
	"github.com/rs/zerolog/log"

	logger "github.com/a-palonskaa/metrics-server/internal/logger"
)

func main() {
	logger.InitLogger("logs/info_agent.log")

	if err := Cmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}
