// Package main provides metrics storage server
// made by @aliffka
package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/reflection" //DEBUG

	proto "github.com/a-palonskaa/metrics-server/gen/proto"
	repo "github.com/a-palonskaa/metrics-server/internal/repository"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	serverREST "github.com/a-palonskaa/metrics-server/internal/server/service/REST"
	serverGRPC "github.com/a-palonskaa/metrics-server/internal/server/service/gRPC"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "a", "a", defaultEndpointAddr, "endpoint HTTP-server adress")
	cmd.PersistentFlags().IntVarP(&Flags.StoreInterval, "i", "i", defaultStoreInterval, "Saving server data interval")
	cmd.PersistentFlags().BoolVarP(&Flags.Restore, "r", "r", defaultRestore, "Saving or not data saved before")
	cmd.PersistentFlags().StringVarP(&Flags.FileStoragePath, "f", "f", defaultFileStoragePath, "Filepath")
	cmd.PersistentFlags().StringVarP(&Flags.DatabaseAddr, "d", "d", defaultDatabaseAddr, "Database filepath")
	cmd.PersistentFlags().StringVarP(&Flags.Key, "k", "k", defaultKey, "Key for hash")
	cmd.PersistentFlags().StringVarP(&Flags.ConfigFile, "config", "c", defaultConfigFile, "Config file")
	cmd.PersistentFlags().StringVarP(&Flags.TrustedSubnet, "t", "t", defaultTrustedSubnet, "Trusted Subnet")
	cmd.PersistentFlags().StringVarP(&Flags.Protocol, "protocol", "p", defaultProtocol, "Protocol(rest/grpc)")
}

var cmd = &cobra.Command{
	Use:   "server",
	Short: "http-server for runtime metrics collection",
	Long: color.New(color.FgGreen).Sprint(`
    	███████╗███████╗██████╗ ██╗   ██╗███████╗██████╗
    	██╔════╝██╔════╝██╔══██╗██║   ██║██╔════╝██╔══██╗
    	███████╗█████╗  ██████╔╝██║   ██║█████╗  ██████╔╝
    	╚════██║██╔══╝  ██╔══██╗╚██╗ ██╔╝██╔══╝  ██╔══██╗
    	███████║███████╗██║  ██║ ╚████╔╝ ███████╗██║  ██║
    	╚══════╝╚══════╝╚═╝  ╚═╝  ╚═══╝  ╚══════╝╚═╝  ╚═╝` + "\n" +
		"\tHTTP server for runtime metrics collection" + "\n\n" +
		"\t\x1b]8;;https://github.com/aliffka\x1b\\" +
		color.New(color.FgCyan).Sprint("@aliffka") +
		"\t\x1b]8;;\x1b\\"),
	PreRun: func(cmd *cobra.Command, args []string) {
		var cfg ParamsConfig
		if err := env.Parse(&cfg); err != nil {
			log.Error().Msg("environment variables parsing error\n")
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
		memStorage := repo.New(repo.NewParams{
			DatabaseAddr:  Flags.DatabaseAddr,
			FilePath:      Flags.FileStoragePath,
			StoreInterval: Flags.StoreInterval,
			Restore:       Flags.Restore,
		})
		defer func() {
			if err := memStorage.Close(); err != nil {
				log.Error().Err(err).Msg("error closing memStorage")
				return
			}
		}()

		connector := database.NewConn(Flags.DatabaseAddr)

		msUsecase := usecase.NewMemStorage(memStorage)
		pingUsecase := usecase.NewPing(connector)

		var err error
		switch Flags.Protocol {
		case restAPI:
			err = runRESTServer(msUsecase, pingUsecase)
		case grpcAPI:
			err = runGRPCServer(msUsecase, pingUsecase)
		default:
			log.Info().Msg("unsupported protocol: " + Flags.Protocol)
		}

		if err != nil {
			log.Error().Err(err).Msg("server error")
		}
	},
}

func runRESTServer(msUsecase usecase.MemStorage, pingUsecase usecase.Ping) error {
	serverHandler := serverREST.New(serverREST.Params{
		MsUsecase:     msUsecase,
		PingUsecase:   pingUsecase,
		Key:           Flags.Key,
		TrustedSubnet: Flags.TrustedSubnet,
	})

	r := chi.NewRouter()
	r = serverHandler.Router(r)

	server := &http.Server{
		Addr:    Flags.EndpointAddr,
		Handler: r,
	}

	serverErr := make(chan error, 1)
	defer close(serverErr)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			serverErr <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	select {
	case err := <-serverErr:
		return err
	case <-sig:
		if err := server.Shutdown(ctx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("shutdown error")
			}
			return err
		}
	}
	return nil
}

func runGRPCServer(msUsecase usecase.MemStorage, pingUsecase usecase.Ping) error {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(serverGRPC.LoggerInterceptor),
		grpc.ChainUnaryInterceptor(serverGRPC.IPValidationInterceptor(Flags.TrustedSubnet)),
	)

	handler := serverGRPC.NewServerHandler(serverGRPC.Params{
		MsUsecase:     msUsecase,
		PingUsecase:   pingUsecase,
		Key:           Flags.Key,
		TrustedSubnet: Flags.TrustedSubnet,
	})

	proto.RegisterMetricsServiceServer(server, handler)
	reflection.Register(server)

	listen, err := net.Listen("tcp", Flags.EndpointAddr)
	if err != nil {
		return err
	}

	serverErr := make(chan error, 1)
	defer close(serverErr)
	go func() {
		if err := server.Serve(listen); err != nil {
			serverErr <- err
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	select {
	case err := <-serverErr:
		return err
	case <-sig:
		server.GracefulStop()
		return nil
	}
}
