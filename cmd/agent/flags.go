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
	defaultEndpointAddr   = "localhost:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2
	defaultKey            = ""
	defaultRateLimit      = 1
	defaultConfigFile     = "" //./internal/configs/agent_config.json
	defaultProtocol       = "rest"
)

type Config struct {
	EndpointAddr   string
	ReportInterval int
	PollInterval   int
	Key            string
	RateLimit      int
	ConfigFile     string
	Protocol       string
}

type ParamsConfig struct {
	EndpointAddr   string `env:"ADDRESS" json:"address"`
	ReportInterval *int   `env:"STORE_INTERVAL" json:"store_interval,omitempty"`
	PollInterval   *int   `env:"POLL_INTERVAL" json:"poll_interval,omitempty"`
	Key            string `env:"KEY" json:"crypto_key"`
	RateLimit      *int   `env:"RATE_LIMIT" json:"rate_limit,omitempty"`
	ConfigFile     string `env:"CONFIG"`
	Protocol       string `env:"PROTOCOL" json:"protocol"`
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

	if cfg.PollInterval != nil {
		Flags.PollInterval = *cfg.PollInterval
	} else if fileCfg.PollInterval != nil && Flags.PollInterval == defaultPollInterval {
		Flags.PollInterval = *fileCfg.PollInterval
	}

	if cfg.ReportInterval != nil {
		Flags.ReportInterval = *cfg.ReportInterval
	} else if fileCfg.ReportInterval != nil && Flags.ReportInterval == defaultReportInterval {
		Flags.ReportInterval = *fileCfg.ReportInterval
	}

	if cfg.Key != "" {
		Flags.Key = cfg.Key
	} else if fileCfg.Key != "" && Flags.Key != defaultKey {
		Flags.Key = fileCfg.Key
	}

	if cfg.RateLimit != nil {
		Flags.RateLimit = *cfg.RateLimit
	} else if fileCfg.RateLimit != nil && Flags.RateLimit == defaultRateLimit {
		Flags.RateLimit = *fileCfg.RateLimit
	}

	if cfg.Protocol != "" {
		Flags.Protocol = cfg.Protocol
	} else if fileCfg.Protocol != "" && Flags.Protocol == defaultProtocol {
		Flags.Protocol = fileCfg.Protocol
	}
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
