package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	ha "github.com/a-palonskaa/metrics-server/internal/handlers/agent"
	mt "github.com/a-palonskaa/metrics-server/internal/metrics"
	st "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

type Config struct {
	EndpointAddr   string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func flagsInit() {
	flag.StringVar(&EndpointAddr, "a", "localhost:8080", "endpoint HTTP-server adress")
	flag.IntVar(&mt.PollInterval, "p", 2, "PollInterval value")
	flag.IntVar(&mt.ReportInterval, "r", 10, "ReportInterval value")

	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Printf("environment variables parsing error\n")
		os.Exit(1)
	}

	if cfg.EndpointAddr != " " {
		EndpointAddr = cfg.EndpointAddr
	}
	if cfg.PollInterval != 0 {
		mt.PollInterval = cfg.PollInterval
	}
	if cfg.ReportInterval != 0 {
		mt.ReportInterval = cfg.PollInterval
	}

	if mt.PollInterval <= 0 || mt.ReportInterval <= 0 {
		fmt.Printf("Error: PollInterval & ReportInterval must be greater than 0\n")
		flag.Usage()
		os.Exit(1)
	}

	parts := strings.Split(EndpointAddr, ":")
	if len(parts) == 1 || len(parts) > 2 {
		fmt.Printf("Error: No port || more than 1 port\n")
		flag.Usage()
		os.Exit(1)
	}

	port, err := strconv.Atoi(parts[1])
	if err != nil || port <= 0 {
		fmt.Printf("Error: Port must be >0 number\n")
		flag.Usage()
		os.Exit(1)
	}
}

var EndpointAddr string

func main() {
	flagsInit()

	memStats := &runtime.MemStats{}

	mt.Update(st.MS, memStats)
	go mt.UpdateRoutine(st.MS, memStats)

	client := &http.Client{}
	for {
		for key, val := range st.MS.GaugeMetrics {
			if err := ha.SendRequest(client, EndpointAddr, "gauge", key, val); err != nil {
				fmt.Printf("Agent: Error sending gauge metric %s: %v\n", key, err)
			}
		}

		for key, val := range st.MS.CounterMetrics {
			if err := ha.SendRequest(client, EndpointAddr, "counter", key, val); err != nil {
				fmt.Printf("Agent: Error sending counter metric %s: %v\n", key, err)
			}
		}
		time.Sleep(time.Duration(mt.ReportInterval) * 1e9)
	}
}
