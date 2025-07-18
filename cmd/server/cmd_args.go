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

	database "github.com/a-palonskaa/metrics-server/internal/database"
	errhandlers "github.com/a-palonskaa/metrics-server/internal/err_handlers"
	server_handler "github.com/a-palonskaa/metrics-server/internal/handlers/server"
	memstorage "github.com/a-palonskaa/metrics-server/internal/metrics_storage"
)

func initMemStorage(ms *memstorage.MemStorage, db *sql.DB, databaseAddr string) {
	if databaseAddr != "" {
		err := errhandlers.RetriableErrHadlerVoid(
			func() error { return database.CreateTables(db) },
			errhandlers.CompareErrSQL)
		if err != nil {
			log.Fatal().Err(err)
		}
		var myDB database.MyDB
		myDB.DB = db
		*ms = myDB
		return
	}
	*ms = memstorage.MS
}

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "a", "a", "localhost:8080", "endpoint HTTP-server adress")
	cmd.PersistentFlags().IntVarP(&Flags.StoreInterval, "i", "i", 300, "Saving server data interval")
	cmd.PersistentFlags().BoolVarP(&Flags.Restore, "r", "r", true, "Saving or not data saved before")
	cmd.PersistentFlags().StringVarP(&Flags.FileStoragePath, "f", "f", "server-data.txt", "Filepath")
	cmd.PersistentFlags().StringVarP(&Flags.DatabaseAddr, "d", "d", "", "Database filepath")
}

func readMetricsStorage(filename string) {
	istream, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal().Err(err)
	}

	if err := memstorage.ReadMetricsStorage(istream); err != nil {
		log.Fatal().Err(err)
	}

	if err := istream.Close(); err != nil {
		log.Error().Err(err)
	}
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
		var ms memstorage.MemStorage
		var db *sql.DB

		db, err := errhandlers.RetriableErrHadler(
			func() (*sql.DB, error) { return sql.Open("pgx", Flags.DatabaseAddr) },
			errhandlers.CompareErrSQL,
		)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to initialize *sql.DB and create a connection pull")
		}
		defer func() {
			if err := db.Close(); err != nil {
				log.Fatal().Err(err)
			}
		}()

		initMemStorage(&ms, db, Flags.DatabaseAddr)

		if Flags.Restore {
			readMetricsStorage(Flags.FileStoragePath)
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

		r = server_handler.RouteRequests(r, db, ms, Flags.StoreInterval, ostream)

		if err := http.ListenAndServe(Flags.EndpointAddr, r); err != nil {
			log.Fatal().Msgf("error loading server: %s", err)
		}
	},
}
