package main

import (
	"net/http"

	hs "github.com/a-palonskaa/metrics-server/internal/handlers/server"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/gauge/`, hs.MakeHandler(hs.GaugeHandler))
	mux.HandleFunc(`/update/counter/`, hs.MakeHandler(hs.CounterHandler))
	mux.HandleFunc(`/`, hs.GeneralCaseHandler)

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
