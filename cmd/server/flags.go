package main

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/rs/zerolog/log"
)

const (
	minPort int = 1
	maxPort int = 65535

	restAPI string = "rest"
	grpcAPI string = "grpc"
)

const (
	defaultEndpointAddr    = "localhost:8080"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "server-data.txt"
	defaultRestore         = true
	defaultDatabaseAddr    = ""
	defaultKey             = ""
	defaultConfigFile      = "" // ./internal/configs/server_config.json
	defaultTrustedSubnet   = ""
	defaultProtocol        = "rest"
)

type Config struct {
	EndpointAddr    string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
	DatabaseAddr    string
	Key             string
	ConfigFile      string
	TrustedSubnet   string
	Protocol        string
}

type ParamsConfig struct {
	EndpointAddr    string `env:"ADDRESS" json:"address"`
	StoreInterval   *int   `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         *bool  `env:"RESTORE" json:"restore,omitempty"`
	DatabaseAddr    string `env:"DATABASE_DSN" json:"database_dns"`
	Key             string `env:"KEY" json:"crypto_key"`
	ConfigFile      string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
	Protocol        string `env:"PROTOCOL" json:"protocol"`
}

var Flags Config

func parseConfigFile(name string, cfg *ParamsConfig) error {
	if name == "" {
		return nil
	}

	file, err := os.OpenFile(name, os.O_RDONLY, 0666)
	if err != nil {
		log.Info().Err(err).Msg("failed to open config file")
		return err
	}

	data, err := io.ReadAll(file)
	if err != nil {
		log.Info().Err(err).Msg("failed to read file")
		return err
	}

	if err = json.Unmarshal(data, cfg); err != nil {
		log.Info().Err(err).Msg("failed to unmarshal config data")
		return err
	}
	return nil
}

func setConfigFile(cfg *ParamsConfig) {
	if cfg.ConfigFile != "" {
		Flags.ConfigFile = cfg.ConfigFile
	}
}

func setFlags(cfg *ParamsConfig, fileCfg *ParamsConfig) {
	if cfg.EndpointAddr != "" {
		Flags.EndpointAddr = cfg.EndpointAddr
	} else if fileCfg.EndpointAddr != "" && Flags.EndpointAddr == defaultEndpointAddr {
		Flags.EndpointAddr = fileCfg.EndpointAddr
	}

	if cfg.FileStoragePath != "" {
		Flags.FileStoragePath = cfg.FileStoragePath
	} else if fileCfg.FileStoragePath != "" && Flags.FileStoragePath == defaultFileStoragePath {
		Flags.FileStoragePath = fileCfg.FileStoragePath
	}

	if cfg.DatabaseAddr != "" {
		Flags.DatabaseAddr = cfg.DatabaseAddr
	} else if fileCfg.DatabaseAddr != "" && Flags.DatabaseAddr == defaultDatabaseAddr {
		Flags.DatabaseAddr = fileCfg.DatabaseAddr
	}

	if cfg.Restore != nil {
		Flags.Restore = *cfg.Restore
	} else if fileCfg.Restore != nil && Flags.Restore {
		Flags.Restore = *fileCfg.Restore
	}

	if cfg.StoreInterval != nil {
		Flags.StoreInterval = *cfg.StoreInterval
	} else if fileCfg.StoreInterval != nil && Flags.StoreInterval == defaultStoreInterval {
		Flags.StoreInterval = *fileCfg.StoreInterval
	}

	if cfg.Key != "" {
		Flags.Key = cfg.Key
	} else if fileCfg.Key != "" && Flags.Key != defaultKey {
		Flags.Key = fileCfg.Key
	}

	if cfg.TrustedSubnet != "" {
		Flags.TrustedSubnet = cfg.TrustedSubnet
	} else if fileCfg.TrustedSubnet != "" && Flags.TrustedSubnet == defaultDatabaseAddr {
		Flags.TrustedSubnet = fileCfg.TrustedSubnet
	}

	if cfg.Protocol != "" {
		Flags.Protocol = cfg.Protocol
	} else if fileCfg.Protocol != "" && Flags.Protocol == defaultProtocol {
		Flags.Protocol = fileCfg.Protocol
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
