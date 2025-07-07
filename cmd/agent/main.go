package main

import (
	"net"
	"runtime"
	"strconv"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	agent_handler "github.com/a-palonskaa/metrics-server/internal/handlers/agent"
	logger "github.com/a-palonskaa/metrics-server/internal/logger"
	metrics "github.com/a-palonskaa/metrics-server/internal/metrics"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

type Config struct {
	EndpointAddr   string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func init() {
	cmd.PersistentFlags().StringVarP(&EndpointAddr, "address", "a", "localhost:8080", "Server endpoint address")
	cmd.PersistentFlags().IntVarP(&metrics.PollInterval, "pollinterval", "p", 2, "Metrics polling interval")
	cmd.PersistentFlags().IntVarP(&metrics.ReportInterval, "reportinterval", "r", 10, "Metrics reporting interval")
}

var EndpointAddr string

var cmd = &cobra.Command{
	Use:   "agent",
	Short: "agent that send runtime metrics to server",
	Long: color.New(color.FgGreen).Sprint(`
    	 █████╗  ██████╗ ███████╗███╗   ██╗████████╗
    	██╔══██╗██╔════╝ ██╔════╝████╗  ██║╚══██╔══╝
    	███████║██║  ███╗█████╗  ██╔██╗ ██║   ██║
    	██╔══██║██║   ██║██╔══╝  ██║╚██╗██║   ██║
    	██║  ██║╚██████╔╝███████╗██║ ╚████║   ██║
    	╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═══╝   ╚═╝`+"\n\n"+
		"\tagent that send runtime metrics to server") + "\n\n" +
		"\t\x1b]8;;https://github.com/aliffka\x1b\\" +
		color.New(color.FgCyan).Sprint("@aliffka") +
		"\t\x1b]8;;\x1b\\",
	PreRun: func(cmd *cobra.Command, args []string) {
		var cfg Config
		err := env.Parse(&cfg)
		if err != nil {
			log.Fatal().Msgf("environment variables parsing error")
		}

		if cfg.EndpointAddr != "" {
			EndpointAddr = cfg.EndpointAddr
		}
		if cfg.PollInterval != 0 {
			metrics.PollInterval = cfg.PollInterval
		}
		if cfg.ReportInterval != 0 {
			metrics.ReportInterval = cfg.PollInterval
		}

		if metrics.PollInterval <= 0 || metrics.ReportInterval <= 0 {
			log.Fatal().Msgf("Error: PollInterval & ReportInterval must be greater than 0")
		}

		_, portStr, err := net.SplitHostPort(EndpointAddr)
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
	},
	Run: func(cmd *cobra.Command, args []string) {
		memStats := &runtime.MemStats{}

		metrics.Update(memstorage.MS, memStats)
		go metrics.UpdateRoutine(memstorage.MS, memStats)

		backoffSchedule := []time.Duration{
			100 * time.Millisecond,
			500 * time.Millisecond,
			1 * time.Second,
		}

		client := resty.New()
		for {
			for key, val := range memstorage.MS.GaugeMetrics {
				for _, backoff := range backoffSchedule {
					err := agent_handler.SendRequest(client, EndpointAddr, "gauge", key, val)
					if err == nil {
						break
					}
					log.Error().Msgf("Agent: Error sending gauge metric %s: %v\n", key, err)
					time.Sleep(backoff)
				}
			}

			for key, val := range memstorage.MS.CounterMetrics {
				for _, backoff := range backoffSchedule {
					err := agent_handler.SendRequest(client, EndpointAddr, "counter", key, val)
					if err == nil {
						break
					}
					log.Error().Msgf("Agent: Error sending counter metric %s: %v\n", key, err)
					time.Sleep(backoff)
				}
			}
			time.Sleep(time.Duration(metrics.ReportInterval) * 1e9)
		}
	},
}

func main() {
	logger.InitLogger("info.log")

	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}
