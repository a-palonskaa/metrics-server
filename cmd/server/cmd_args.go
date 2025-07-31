// Package main provides metrics storage server
// made by @aliffka
package main

import (
	"context"
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

	repo "github.com/a-palonskaa/metrics-server/internal/repository"
	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	service "github.com/a-palonskaa/metrics-server/internal/server/service/REST"
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
	cmd.PersistentFlags().StringVarP(&Flags.TrustedSubnet, "t", "t", defaultTrustedSubnet, "Trusted Sunnet")
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
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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

		serverHandler := service.New(service.Params{
			MsUsecase:     msUsecase,
			PingUsecase:   pingUsecase,
			Key:           Flags.Key,
			TrustedSubnet: Flags.TrustedSubnet,
		})

		r := chi.NewRouter()
		r = serverHandler.Router(r)

		serverErr := make(chan error, 1)
		defer close(serverErr)

		server := &http.Server{
			Addr:    Flags.EndpointAddr,
			Handler: r,
		}

		go func() {
			if err := server.ListenAndServe(); err != nil {
				serverErr <- err
				return
			}
		}()

		select {
		case err := <-serverErr:
			log.Error().Err(err).Msg("server error")
			cancel()
			return
		case <-sig:
			if err := server.Shutdown(ctx); err != nil {
				if closeErr := server.Close(); closeErr != nil {
					log.Error().Err(closeErr).Msg("shutdown error")
				}
			}
		}
	},
}
