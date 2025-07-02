package main

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"

	hs "github.com/a-palonskaa/metrics-server/internal/handlers/server"
)

type Config struct {
	EndpointAddr string `env:"ADDRESS"`
}

var EndpointAddr string

func main() {
	flag.StringVar(&EndpointAddr, "a", "localhost:8080", "endpoint HTTP-server adress")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		fmt.Printf("environment variables parsing error\n")
		os.Exit(1)
	}

	if cfg.EndpointAddr != " " {
		EndpointAddr = cfg.EndpointAddr
	}

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

	http.ListenAndServe(EndpointAddr, r)
}
