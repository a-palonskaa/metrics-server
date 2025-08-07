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

	"github.com/fatih/color"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/encoding/gzip"

	proto "github.com/a-palonskaa/metrics-server/gen/proto"
	service "github.com/a-palonskaa/metrics-server/internal/agent/service"
	agentrest "github.com/a-palonskaa/metrics-server/internal/agent/service/REST"
	agentgrpc "github.com/a-palonskaa/metrics-server/internal/agent/service/gRPC"
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
	cmd.PersistentFlags().StringVarP(&Flags.Protocol, "protocol", "t", defaultProtocol, "Protocol(rest/grpc)")
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
		v := viper.New()

		if err := initConfig(v, Flags.ConfigFile); err != nil {
			log.Error().Err(err).Msg("failed to load config")
		}

		_ = v.BindPFlag("address", cmd.PersistentFlags().Lookup("address"))
		_ = v.BindPFlag("poll_interval", cmd.PersistentFlags().Lookup("pollinterval"))
		_ = v.BindPFlag("report_interval", cmd.PersistentFlags().Lookup("reportinterval"))
		_ = v.BindPFlag("key", cmd.PersistentFlags().Lookup("key"))
		_ = v.BindPFlag("rate_limit", cmd.PersistentFlags().Lookup("ratelimit"))
		_ = v.BindPFlag("protocol", cmd.PersistentFlags().Lookup("protocol"))

		Flags = Config{
			EndpointAddr:   v.GetString("address"),
			PollInterval:   v.GetInt("poll_interval"),
			ReportInterval: v.GetInt("report_interval"),
			Key:            v.GetString("key"),
			RateLimit:      v.GetInt("rate_limit"),
			Protocol:       v.GetString("protocol"),
		}
		validateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		var handler service.Handler
		switch Flags.Protocol {
		case restAPI:
			client := resty.New().SetBaseURL("http://" + Flags.EndpointAddr)
			handler = agentrest.NewHandler(memstorage.New(), Flags.Key, client)
		case grpcAPI:
			conn, err := grpc.NewClient(Flags.EndpointAddr,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")),
				grpc.WithChainUnaryInterceptor(agentgrpc.HashSigning(Flags.Key)),
			)
			if err != nil {
				log.Fatal().Err(err).Msg("failed to establish gRPC conn")
			}
			defer func() {
				if err := conn.Close(); err != nil {
					log.Info().Err(err).Msg("failed to close connection")
				}
			}()

			client := proto.NewMetricsServiceClient(conn)
			handler = agentgrpc.NewHandler(memstorage.New(), client)
		default:
			log.Info().Msg("unallowed protocol")
			return
		}

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(sig)

		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				log.Error().Err(err).Msg("pprof server error")
			}
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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
						return handler.SendMetrics(ctx)
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
