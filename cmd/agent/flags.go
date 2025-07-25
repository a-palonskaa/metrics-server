package main

import (
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	minPort int = 1
	maxPort int = 65535
)

type Config struct {
	EndpointAddr   string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

var Flags Config

func setFlags(cfg *Config) {
	if cfg.EndpointAddr != "" {
		Flags.EndpointAddr = cfg.EndpointAddr
	}
	if cfg.PollInterval != 0 {
		Flags.PollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval != 0 {
		Flags.ReportInterval = cfg.PollInterval
	}
}

func validateFlags() {
	if Flags.PollInterval <= 0 || Flags.ReportInterval <= 0 {
		log.Error().Msgf("Error: PollInterval & ReportInterval must be greater than 0")
		return
	}

	_, portStr, err := net.SplitHostPort(Flags.EndpointAddr)
	if err != nil {
		log.Error().Msgf("invalid address format: %s", err)
		return
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Error().Msgf("port must be a number: %s", err)
		return
	}

	if port < minPort || port > maxPort {
		log.Error().Msgf("port must be between %d and %d", minPort, maxPort)
		return
	}
}
