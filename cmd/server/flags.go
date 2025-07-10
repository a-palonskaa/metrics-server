package main

import (
	"net"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

type Config struct {
	EndpointAddr    string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
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

func setFlags(cfg *Config) {
	if cfg.EndpointAddr != "" {
		Flags.EndpointAddr = cfg.EndpointAddr
	}

	if cfg.FileStoragePath != "" {
		Flags.FileStoragePath = cfg.FileStoragePath
	}

	if _, exists := os.LookupEnv("RESTORE"); exists {
		Flags.Restore = cfg.Restore
	}

	if _, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		Flags.StoreInterval = cfg.StoreInterval
	}
}
