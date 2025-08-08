package main

import (
	"fmt"
	"net"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
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
	defaultConfigFile      = "" // ./internal/configs/server_config.yaml
	defaultTrustedSubnet   = ""
	defaultProtocol        = "rest"
)

type Config struct {
	EndpointAddr    string `env:"ADDRESS" yaml:"address"`
	StoreInterval   int    `env:"STORE_INTERVAL" yaml:"store_interval,omitempty"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" yaml:"store_file"`
	Restore         bool   `env:"RESTORE" yaml:"restore,omitempty"`
	DatabaseAddr    string `env:"DATABASE_DSN" yaml:"database_dsn"`
	Key             string `env:"KEY" yaml:"key"`
	ConfigFile      string `env:"CONFIG"`
	TrustedSubnet   string `env:"TRUSTED_SUBNET" yaml:"trusted_subnet"`
	Protocol        string `env:"PROTOCOL" yaml:"protocol"`
}

var Flags Config

func initConfig(v *viper.Viper, configFile string) error {
	v.AutomaticEnv()

	v.SetDefault("address", defaultEndpointAddr)
	v.SetDefault("store_interval", defaultStoreInterval)
	v.SetDefault("store_file", defaultFileStoragePath)
	v.SetDefault("restore", defaultRestore)
	v.SetDefault("database_dsn", defaultDatabaseAddr)
	v.SetDefault("key", defaultKey)
	v.SetDefault("trusted_subnet", defaultTrustedSubnet)
	v.SetDefault("protocol", defaultProtocol)

	cfgFile := configFile
	if cfgFile == "" && v.GetString("config") != "" {
		cfgFile = v.GetString("config")
	}

	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
		if err := v.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}
	return nil
}

func validateFlags() {
	_, portStr, err := net.SplitHostPort(Flags.EndpointAddr)
	if err != nil {
		log.Fatal().Msgf("invalid address %s format: %s", Flags.EndpointAddr, err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal().Msgf("port must be a number: %s", err)
	}

	if port < minPort || port > maxPort {
		log.Fatal().Msgf("port must be between %d and %d", minPort, maxPort)
	}
}
