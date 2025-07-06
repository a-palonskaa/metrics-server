package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/fatih/color"
	"github.com/go-chi/chi/v5"
	"github.com/spf13/cobra"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	hs "github.com/a-palonskaa/metrics-server/internal/handlers/server"
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
			log.Printf("environment variables parsing error\n")
			os.Exit(1)
		}

		if cfg.EndpointAddr != "" {
			EndpointAddr = cfg.EndpointAddr
		}

		_, portStr, err := net.SplitHostPort(EndpointAddr)
		if err != nil {
			log.Printf("invalid address format: %s", err)
			os.Exit(1)
		}

		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("port must be a number: %s", err)
			os.Exit(1)
		}

		if port < 1 || port > 65535 {
			log.Printf("port must be between 1 and 65535")
			os.Exit(1)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		r := chi.NewRouter()

		r.Route("/value", func(r chi.Router) {
			r.Get("/", hs.AllValueHandler)
			r.Route("/gauge", func(r chi.Router) {
				r.Get("/", hs.NoNameHandler)
				r.Get("/{name}", hs.GaugeGetHandler)
			})
			r.Route("/counter", func(r chi.Router) {
				r.Get("/", hs.NoNameHandler)
				r.Get("/{name}", hs.CounterGetHandler)
			})
		})

		r.Route("/update", func(r chi.Router) {
			r.Route("/gauge", func(r chi.Router) {
				r.Post("/", hs.NoNameHandler)
				r.Route("/{name}", func(r chi.Router) {
					r.Post("/*", hs.NoValueHandler)
					r.Post("/{value}", hs.GaugePostHandler)
				})
			})
			r.Route("/counter", func(r chi.Router) {
				r.Post("/", hs.NoNameHandler)
				r.Route("/{name}", func(r chi.Router) {
					r.Post("/*", hs.NoValueHandler)
					r.Post("/{value}", hs.CounterPostHandler)
				})
			})
			r.Post("/*", hs.GeneralCaseHandler)
		})
		r.Handle("/", http.HandlerFunc(hs.GeneralCaseHandler))

		if err := http.ListenAndServe(EndpointAddr, r); err != nil {
			log.Fatalf("error loading server: %s", err)
		}
	},
}

func init() {
	cmd.PersistentFlags().StringVarP(&EndpointAddr, "address", "a", "localhost:8080", "endpoint HTTP-server adress")
}

func main() {
	if err := cmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
