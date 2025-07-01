package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"

	hs "github.com/a-palonskaa/metrics-server/internal/handlers/server"
)

func main() {
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
	})

	r.Handle("/", http.HandlerFunc(hs.GeneralCaseHandler))

	http.ListenAndServe(":8080", r)
}
