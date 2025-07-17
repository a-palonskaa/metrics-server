package main

import (
	"net/http"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	server_handler "github.com/a-palonskaa/metrics-server/internal/handlers/server"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "a", "a", "localhost:8080", "endpoint HTTP-server adress")
	cmd.PersistentFlags().IntVarP(&Flags.StoreInterval, "i", "i", 300, "Saving server data interval")
	cmd.PersistentFlags().BoolVarP(&Flags.Restore, "r", "r", true, "Saving or not data saved before")
	cmd.PersistentFlags().StringVarP(&Flags.FileStoragePath, "f", "f", "server-data.txt", "Filepath")
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
		var cfg Config
		if err := env.Parse(&cfg); err != nil {
			log.Fatal().Msgf("environment variables parsing error\n")
		}

		setFlags(&cfg)
		validateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		istream, err := os.OpenFile(Flags.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal().Err(err)
		}

		if Flags.Restore {
			if err := memstorage.ReadMetricsStorage(istream); err != nil {
				log.Fatal().Err(err)
			}
		}

		if err := istream.Close(); err != nil {
			log.Error().Err(err)
		}

		ostream, err := os.OpenFile(Flags.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatal().Err(err)
		}
		defer func() {
			if err := ostream.Close(); err != nil {
				log.Error().Err(err)
			}
		}()

		r := chi.NewRouter()

		r.Use(server_handler.WithCompression)
		r.Use(server_handler.WithLogging)

		if Flags.StoreInterval == 0 {
			r.Use(server_handler.MakeSavingHandler(ostream))
		} else {
			memstorage.RunSavingStorageRoutine(ostream, Flags.StoreInterval)
		}

		server_handler.RouteRequests(r)

		if err := http.ListenAndServe(Flags.EndpointAddr, r); err != nil {
			log.Fatal().Msgf("error loading server: %s", err)
		}
	},
}
