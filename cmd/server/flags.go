package main

import (
	"net"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	minPort int = 1
	maxPort int = 65535
)

type Config struct {
	EndpointAddr    string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseAddr    string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

var Flags Config

func setFlags(cfg *Config) {
	if cfg.EndpointAddr != "" {
		Flags.EndpointAddr = cfg.EndpointAddr
	}

	if cfg.FileStoragePath != "" {
		Flags.FileStoragePath = cfg.FileStoragePath
	}

	if cfg.DatabaseAddr != "" {
		Flags.DatabaseAddr = cfg.DatabaseAddr
	}

	if _, exists := os.LookupEnv("RESTORE"); exists {
		Flags.Restore = cfg.Restore
	}

	if _, exists := os.LookupEnv("STORE_INTERVAL"); exists {
		Flags.StoreInterval = cfg.StoreInterval
	}

	if cfg.Key != "" {
		Flags.Key = cfg.Key
	}
}

func validateFlags() {
	_, portStr, err := net.SplitHostPort(Flags.EndpointAddr)
	if err != nil {
		log.Fatal().Msgf("invalid address format: %s", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal().Msgf("port must be a number: %s", err)
	}

	if port < minPort || port > maxPort {
		log.Fatal().Msgf("port must be between %d and %d", minPort, maxPort)
	}
}
