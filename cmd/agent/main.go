package main

import (
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"

	agent_handler "github.com/a-palonskaa/metrics-server/internal/handlers/agent"
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
			fmt.Printf("environment variables parsing error\n")
			os.Exit(1)
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
			log.Printf("Error: PollInterval & ReportInterval must be greater than 0\n")
			os.Exit(1)
		}

		_, portStr, err := net.SplitHostPort(EndpointAddr)
		if err != nil {
			log.Printf("invalid address format: %s", err)
			os.Exit(1)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("port must be a number: %s", err)
			os.Exit(1)
		}

		if port < 1 || port > 65535 {
			log.Printf("port must be between 1 and 65535")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		memStats := &runtime.MemStats{}

		metrics.Update(memstorage.MS, memStats)
		go metrics.UpdateRoutine(memstorage.MS, memStats)

		client := resty.New()
		for {
			for key, val := range memstorage.MS.GaugeMetrics {
				if err := agent_handler.SendRequest(client, EndpointAddr, "gauge", key, val); err != nil {
					fmt.Printf("Agent: Error sending gauge metric %s: %v\n", key, err)
				}
			}

			for key, val := range memstorage.MS.CounterMetrics {
				if err := agent_handler.SendRequest(client, EndpointAddr, "counter", key, val); err != nil {
					fmt.Printf("Agent: Error sending counter metric %s: %v\n", key, err)
				}
			}
			time.Sleep(time.Duration(metrics.ReportInterval) * 1e9)
		}
	},
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
