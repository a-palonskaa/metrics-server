package main

import (
	"fmt"
	"runtime"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	agent_handler "github.com/a-palonskaa/metrics-server/internal/handlers/agent"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func init() {
	Cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "address", "a", "localhost:8080", "Server endpoint address")
	Cmd.PersistentFlags().IntVarP(&Flags.PollInterval, "pollinterval", "p", 2, "Metrics polling interval")
	Cmd.PersistentFlags().IntVarP(&Flags.ReportInterval, "reportinterval", "r", 10, "Metrics reporting interval")
}

var Cmd = &cobra.Command{
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
			Flags.EndpointAddr = cfg.EndpointAddr
		}
		if cfg.PollInterval != 0 {
			Flags.PollInterval = cfg.PollInterval
		}
		if cfg.ReportInterval != 0 {
			Flags.ReportInterval = cfg.PollInterval
		}

		validateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		memStats := &runtime.MemStats{}

		memstorage.MS.Update(memStats)
		go memstorage.MS.UpdateRoutine(memStats, time.Duration(Flags.PollInterval)*1e9)

		backoffSchedule := []time.Duration{
			100 * time.Millisecond,
			500 * time.Millisecond,
			1 * time.Second,
		}

		client := resty.New()
		for {
			memstorage.MS.Iterate(func(key string, mType string, val fmt.Stringer) {
				for _, backoff := range backoffSchedule {
					err := agent_handler.SendRequest(client, Flags.EndpointAddr, mType, key, val)
					if err == nil {
						break
					}
					log.Error().Msgf("error sending %s metric %s(%v): %v\n", mType, key, val, err)
					time.Sleep(backoff)
				}
			})
			time.Sleep(time.Duration(Flags.ReportInterval) * 1e9)
		}
	},
}
