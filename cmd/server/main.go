package main

import (
	"github.com/rs/zerolog/log"

	logger "github.com/a-palonskaa/metrics-server/pkg/logger"
)

// @Title MetricsServer API
// @Description metrics storage server
// @Version 1.0
// @Contact.email polonskaia.aa@phystech.edu
// @BasePath /
// @Host localhost:8080

func main() {
	logger.InitLogger("logs/info.log")

	if err := cmd.Execute(); err != nil {
		log.Error().Err(err)
	}
}
