// Package main provides metrics storage server
// made by @aliffka
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/rookie-ninja/rk-boot"
	"github.com/rookie-ninja/rk-grpc/boot"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"

	proto "github.com/a-palonskaa/metrics-server/gen/proto"
	repo "github.com/a-palonskaa/metrics-server/internal/repository"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	serverrest "github.com/a-palonskaa/metrics-server/internal/server/service/REST"
	servergrpc "github.com/a-palonskaa/metrics-server/internal/server/service/gRPC"
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
		v := viper.New()
		if err := initConfig(v, Flags.ConfigFile); err != nil {
			log.Error().Err(err).Msg("failed to load config")
		}

		_ = v.BindPFlag("address", cmd.PersistentFlags().Lookup("a"))
		_ = v.BindPFlag("store_interval", cmd.PersistentFlags().Lookup("i"))
		_ = v.BindPFlag("store_file", cmd.PersistentFlags().Lookup("f"))
		_ = v.BindPFlag("restore", cmd.PersistentFlags().Lookup("r"))
		_ = v.BindPFlag("database_dsn", cmd.PersistentFlags().Lookup("d"))
		_ = v.BindPFlag("key", cmd.PersistentFlags().Lookup("k"))
		_ = v.BindPFlag("trusted_subnet", cmd.PersistentFlags().Lookup("t"))
		_ = v.BindPFlag("protocol", cmd.PersistentFlags().Lookup("protocol"))

		Flags = Config{
			EndpointAddr:    v.GetString("address"),
			StoreInterval:   v.GetInt("store_interval"),
			FileStoragePath: v.GetString("store_file"),
			Restore:         v.GetBool("restore"),
			DatabaseAddr:    v.GetString("database_dsn"),
			Key:             v.GetString("key"),
			TrustedSubnet:   v.GetString("trusted_subnet"),
			Protocol:        v.GetString("protocol"),
		}
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
			if memStorage != nil {
				if err := memStorage.Close(); err != nil {
					log.Error().Err(err).Msg("error closing memStorage")
					return
				}
			}
		}()

		log.Info().Msgf("Database addr is %s", Flags.DatabaseAddr)
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
	log.Info().Msgf("%s\n%s\n%s", Flags.EndpointAddr, Flags.Key)
	serverHandler := serverrest.New(serverrest.Params{
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
		return fmt.Errorf("server error:%w", err)
	case <-sig:
		if err := server.Shutdown(ctx); err != nil {
			if closeErr := server.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg("shutdown error")
			}
			return fmt.Errorf("faile dto gracefully shut down:%w", err)
		}
	}
	return nil
}

func runGRPCServer(msUsecase usecase.MemStorage, pingUsecase usecase.Ping) error {
	boot := rkboot.NewBoot(rkboot.WithBootConfigPath(""))
	boot.Bootstrap(context.Background())

	grpcEntry := rkgrpc.GetGrpcEntry("metrics-server")
	grpcEntry.Port = getPort(Flags.EndpointAddr)
	grpcEntry.AddRegFuncGrpc(func(server *grpc.Server) {
		proto.RegisterMetricsServiceServer(server, servergrpc.NewServerHandler(servergrpc.Params{
			MsUsecase:     msUsecase,
			PingUsecase:   pingUsecase,
			Key:           Flags.Key,
			TrustedSubnet: Flags.TrustedSubnet,
		}))
	})
	grpcEntry.AddServerOptions(
		grpc.ChainUnaryInterceptor(
			servergrpc.LoggerInterceptor,
			servergrpc.IPValidationInterceptor(Flags.TrustedSubnet),
		),
	)

	boot.WaitForShutdownSig(context.Background())
	return nil
}

func getPort(endpointAddr string) uint64 {
	_, portStr, err := net.SplitHostPort(endpointAddr)
	if err != nil {
		log.Info().Err(err).Msgf("failed to sptil host addr: %s", endpointAddr)
		return 0
	}

	port, err := strconv.ParseInt(portStr, 0, 64)
	if err != nil {
		log.Info().Err(err).Msgf("failed to converte %s to type int", portStr)
	}
	return uint64(port)
}
