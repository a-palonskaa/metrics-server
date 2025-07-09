package main

import (
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
)

type Config struct {
	EndpointAddr string `env:"ADDRESS"`
}

var Flags Config

func validateFlags() {
	_, portStr, err := net.SplitHostPort(Flags.EndpointAddr)
	if err != nil {
		log.Fatal().Msgf("invalid address format: %s", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal().Msgf("port must be a number: %s", err)
	}

	if port < 1 || port > 65535 {
		log.Fatal().Msgf("port must be between 1 and 65535")
	}
}
