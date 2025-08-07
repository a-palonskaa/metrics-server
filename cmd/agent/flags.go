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
	defaultEndpointAddr   = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultKey            = ""
	defaultRateLimit      = 1
	defaultConfigFile     = "" //./internal/configs/agent_config.yaml
	defaultProtocol       = "rest"
)

type Config struct {
	EndpointAddr   string `env:"ADDRESS" json:"address"`
	ReportInterval int    `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	PollInterval   int    `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	Key            string `env:"KEY" json:"crypto_key"`
	RateLimit      int    `env:"RATE_LIMIT" json:"rate_limit,omitempty"`
	ConfigFile     string `env:"CONFIG"`
	Protocol       string `env:"PROTOCOL" json:"protocol"`
}

var Flags Config

func initConfig(v *viper.Viper, configFile string) error {
	v.AutomaticEnv()

	v.SetDefault("endpointaddr", defaultEndpointAddr)
	v.SetDefault("reportinterval", defaultReportInterval)
	v.SetDefault("pollinterval", defaultPollInterval)
	v.SetDefault("ratelimit", defaultRateLimit)
	v.SetDefault("key", defaultKey)
	v.SetDefault("config", defaultConfigFile)
	v.SetDefault("protocol", defaultProtocol)

	cfgFile := configFile
	if cfgFile == "" && v.GetString("configure") != "" {
		cfgFile = v.GetString("configure")
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
	if Flags.PollInterval <= 0 || Flags.ReportInterval <= 0 {
		log.Fatal().Msgf("Error: PollInterval & ReportInterval must be greater than 0")
	}

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
