package main

import (
	"net/http"

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
	service "github.com/a-palonskaa/metrics-server/internal/server/service"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "a", "a", "localhost:8080", "endpoint HTTP-server adress")
	cmd.PersistentFlags().IntVarP(&Flags.StoreInterval, "i", "i", 300, "Saving server data interval")
	cmd.PersistentFlags().BoolVarP(&Flags.Restore, "r", "r", true, "Saving or not data saved before")
	cmd.PersistentFlags().StringVarP(&Flags.FileStoragePath, "f", "f", "server-data.txt", "Filepath")
	cmd.PersistentFlags().StringVarP(&Flags.DatabaseAddr, "d", "d", "", "Database filepath")
	cmd.PersistentFlags().StringVarP(&Flags.Key, "key", "k", "", "Key for hash")
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
			log.Error().Msg("environment variables parsing error\n")
			return
		}

		setFlags(&cfg)
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

		serverHandler := service.New(service.Params{
			MsUsecase:   msUsecase,
			PingUsecase: pingUsecase,
			Key:         Flags.Key,
		})

		r := chi.NewRouter()
		r = serverHandler.Router(r)

		if err := http.ListenAndServe(Flags.EndpointAddr, r); err != nil {
			log.Error().Msgf("error loading server: %s", err)
			return
		}
	},
}
