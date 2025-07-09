package main

import (
	"net/http"

	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	server_handler "github.com/a-palonskaa/metrics-server/internal/handlers/server"
)

func init() {
	cmd.PersistentFlags().StringVarP(&Flags.EndpointAddr, "address", "a", "localhost:8080", "endpoint HTTP-server adress")
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

		if cfg.EndpointAddr != "" {
			Flags.EndpointAddr = cfg.EndpointAddr
		}

		validateFlags()
	},
	Run: func(cmd *cobra.Command, args []string) {
		r := chi.NewRouter()

		r.Route("/value", func(r chi.Router) {
			r.Post("/", server_handler.WithLogging(server_handler.PostJSONValueHandler))
			r.Get("/", server_handler.WithLogging(server_handler.AllValueHandler))
			r.Get("/{mType}/{name}", server_handler.WithLogging(server_handler.GetHandler))
		})
		r.Route("/update", func(r chi.Router) {
			r.Post("/", server_handler.WithLogging(server_handler.PostJSONUpdateHandler))
			r.Post("/{mType}/{name}/{value}", server_handler.WithLogging(server_handler.PostHandler))
		})
		if err := http.ListenAndServe(Flags.EndpointAddr, r); err != nil {
			log.Fatal().Msgf("error loading server: %s", err)
		}
	},
}
