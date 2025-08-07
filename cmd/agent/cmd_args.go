// Package main provides agent that send runtime and system metrics to server
// made by @aliffka

package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	agent_handler "github.com/a-palonskaa/metrics-server/internal/agent/service"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	workerpool "github.com/a-palonskaa/metrics-server/pkg/worker_pool"
)

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "address", "a", defaultEndpointAddr, "Server endpoint address")
	cmd.PersistentFlags().IntVarP(&Flags.PollInterval, "pollinterval", "p", defaultPollInterval, "Metrics polling interval")
	cmd.PersistentFlags().IntVarP(&Flags.ReportInterval, "reportinterval", "r", defaultReportInterval, "Metrics reporting interval")
	cmd.PersistentFlags().StringVarP(&Flags.Key, "key", "k", defaultKey, "Key for hash")
	cmd.PersistentFlags().IntVarP(&Flags.RateLimit, "limit", "l", defaultRateLimit, "Limit for requests amount")
	cmd.PersistentFlags().StringVarP(&Flags.ConfigFile, "config", "c", defaultConfigFile, "Config file")
}

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
		var cfg ParamsConfig
		err := env.Parse(&cfg)
		if err != nil {
			log.Error().Msgf("environment variables parsing error")
			return
		}
		setConfigFile(&cfg)

		var fileCfg ParamsConfig
		if err := parseConfigFile(Flags.ConfigFile, &fileCfg); err != nil {
			log.Error().Msg("config file parsing error\n")
			return
		}

		setFlags(&cfg, &fileCfg)
		validateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				log.Error().Err(err).Msg("pprof server error")
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		client := resty.New().SetBaseURL("http://" + Flags.EndpointAddr)
		handler := agent_handler.NewHandler(memstorage.New())

		updateTicker := time.NewTicker(time.Duration(Flags.PollInterval) * time.Second)
		defer updateTicker.Stop()
		sendTicker := time.NewTicker(time.Duration(Flags.ReportInterval) * time.Second)
		defer sendTicker.Stop()

		w := workerpool.New(Flags.RateLimit, ctx)
		defer w.Close()

		go func() {
			for {
				select {
				case <-updateTicker.C:
					handler.UpdateRuntimeMetrics(ctx)
					handler.UpdateSystemMetrics(ctx)
				case <-sendTicker.C:
					w.AddTask(func(c context.Context) error {
						return handler.SendMetrics(c, client, Flags.Key)
					})
				case <-ctx.Done():
					return
				}
			}
		}()

		go func() {
			for {
				select {
				case err := <-w.Result():
					if err != nil {
						log.Error().Err(err).Msg("failed to send metric")
					}
				case <-ctx.Done():
					return
				}
			}
		}()

		<-sig
		cancel()
	},
}
