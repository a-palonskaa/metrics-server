package main

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

<<<<<<< HEAD
	database "github.com/a-palonskaa/metrics-server/internal/database"
	errhandlers "github.com/a-palonskaa/metrics-server/internal/err_handlers"
	server_handler "github.com/a-palonskaa/metrics-server/internal/handlers/server"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
=======
	repo "github.com/a-palonskaa/metrics-server/internal/repository"
	file "github.com/a-palonskaa/metrics-server/internal/repository/file"
	service "github.com/a-palonskaa/metrics-server/internal/server/service"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
<<<<<<< HEAD
>>>>>>> a77deee (file logic)
=======
>>>>>>> a83d27f (naming+html template)
)

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "a", "a", "localhost:8080", "endpoint HTTP-server adress")
	cmd.PersistentFlags().IntVarP(&Flags.StoreInterval, "i", "i", 300, "Saving server data interval")
	cmd.PersistentFlags().BoolVarP(&Flags.Restore, "r", "r", true, "Saving or not data saved before")
	cmd.PersistentFlags().StringVarP(&Flags.FileStoragePath, "f", "f", "server-data.txt", "Filepath")
	cmd.PersistentFlags().StringVarP(&Flags.DatabaseAddr, "d", "d", "", "Database filepath")
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
<<<<<<< HEAD
		memStorage := repo.NewMemStorage(Flags.DatabaseAddr)
		backupStorage := file.NewFileBackup(Flags.FileStoragePath)

		msUsecase := usecase.NewMemStorageUsecase(memStorage, backupStorage, Flags.StoreInterval, Flags.Restore)
		pingUsecase := usecase.NewPingUsecase(Flags.DatabaseAddr)

		serverHandler := service.NewHandler(msUsecase, pingUsecase)
=======
		repoParams := repo.NewParams{
			DatabaseAddr:  Flags.DatabaseAddr,
			FilePath:      Flags.FileStoragePath,
			StoreInterval: Flags.StoreInterval,
			Restore:       Flags.Restore,
		}
		memStorage := repo.New(repoParams)
>>>>>>> a83d27f (naming+html template)
		defer func() {
			if err := serverHandler.Close(); err != nil {
				log.Error().Err(err).Msg("error closing handler")
				return
			}
		}()

<<<<<<< HEAD
=======
		connector := conn.New(Flags.DatabaseAddr)

		msUsecase := usecase.NewMemStorage(memStorage)
		pingUsecase := usecase.NewPing(connector)

		serverHandler := service.New(msUsecase, pingUsecase)

>>>>>>> a83d27f (naming+html template)
		r := chi.NewRouter()
		r = serverHandler.Router(r)

		if err := http.ListenAndServe(Flags.EndpointAddr, r); err != nil {
			log.Fatal().Msgf("error loading server: %s", err)
		}
	},
}
