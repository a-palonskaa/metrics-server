package main

import (
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
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

		setFlags(&cfg)
		validateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		memStats := &runtime.MemStats{}
		client := resty.New()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		tickerUpdate := time.NewTicker(time.Duration(Flags.PollInterval) * time.Second)
		defer tickerUpdate.Stop()
		tickerSend := time.NewTicker(time.Duration(Flags.ReportInterval) * time.Second)
		defer tickerSend.Stop()

		for {
			select {
			case <-tickerUpdate.C:
				memstorage.MS.Update(ctx, memStats)
			case <-tickerSend.C:
				agent_handler.SendMetrics(ctx, client, Flags.EndpointAddr)
			case <-sig:
				return
			}
		}
	},
}
