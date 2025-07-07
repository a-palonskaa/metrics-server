package main

import (
	"net"
	"net/http"
	"strconv"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	server_handler "github.com/a-palonskaa/metrics-server/internal/handlers/server"
	logger "github.com/a-palonskaa/metrics-server/internal/logger"
)

type Config struct {
	EndpointAddr string `env:"ADDRESS"`
}

var EndpointAddr string

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

		if cfg.EndpointAddr != "" {
			EndpointAddr = cfg.EndpointAddr
		}

		_, portStr, err := net.SplitHostPort(EndpointAddr)
		if err != nil {
			log.Fatal().Msgf("invalid address format: %s", err)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatal().Msgf("port must be a number: %s", err)
		}

		if port < 1 || port > 65535 {
			log.Fatal().Msgf("port must be between 1 and 65535")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		r := chi.NewRouter()

		r.Route("/value", func(r chi.Router) {
			r.Get("/", server_handler.WithLogging(server_handler.AllValueHandler))
			r.Get("/{mType}/{name}", server_handler.WithLogging(server_handler.GetHandler))
		})
		r.Post("/update/{mType}/{name}/{value}", server_handler.WithLogging(server_handler.PostHandler))

		if err := http.ListenAndServe(EndpointAddr, r); err != nil {
			log.Fatal().Msgf("error loading server: %s", err)
		}
	},
}

func init() {
	cmd.PersistentFlags().StringVarP(&EndpointAddr, "address", "a", "localhost:8080", "endpoint HTTP-server adress")
}

func main() {
	logger.InitLogger("info.log")

	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}
