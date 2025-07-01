package main

import (
	"net/http"

	hs "github.com/a-palonskaa/metrics-server/internal/handlers/server"
	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()

	r.Route("/value", func(r chi.Router) {
		r.Get("/", hs.AllValueHandler)
		r.Get("/gauge/", hs.MakeGetHandler(hs.GaugeValueHandler))
		r.Get("/counter/", hs.MakeGetHandler(hs.CounterValueHandler))
	})

	r.Route("/update", func(r chi.Router) {
		r.Post("/gauge/", hs.MakePostHandler(hs.GaugeHandler))
		r.Post("/counter/", hs.MakePostHandler(hs.CounterHandler))
	})

	r.Handle("/", http.HandlerFunc(hs.GeneralCaseHandler))

	http.ListenAndServe(":8080", r)
}
