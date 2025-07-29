package server_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	database "github.com/a-palonskaa/metrics-server/internal/repository/database"
	memstorage "github.com/a-palonskaa/metrics-server/internal/repository/metrics_storage"
	server "github.com/a-palonskaa/metrics-server/internal/server/service"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

// Example_postUpdateMetric demonstrates sending and retrieving a metric
// using the HTTP endpoints of the metrics server.
func Example_postUpdateMetric() {
	// Set up router and in-memory dependencies
	r := chi.NewRouter()

	msUsecase := usecase.NewMemStorage(memstorage.New())
	pingUsecase := usecase.NewPing(database.NewConn(""))

	serverHandler := server.New(server.Params{
		MsUsecase:   msUsecase,
		PingUsecase: pingUsecase,
	})

	_ = serverHandler.Router(r)

	// Launch test HTTP server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Send metric via POST /update/
	updateBody := `{"id": "Alloc", "type": "gauge", "value": 123.4}`
	resp, err := http.Post(ts.URL+"/update/", "application/json", bytes.NewBufferString(updateBody))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to lcose response body: %s", err)
		}
	}()
	fmt.Println("POST /update/:", resp.Status)

	// Retrieve metric via GET /value/gauge/Alloc
	resp, err = http.Get(ts.URL + "/value/gauge/Alloc")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to lcose response body: %s", err)
		}
	}()

	getRespBody, _ := io.ReadAll(resp.Body)
	fmt.Println("GET /value/gauge/Alloc:", string(getRespBody))

	// Request metric via POST /value/
	jsonRequest := `{"id": "Alloc", "type": "gauge"}`
	resp, err = http.Post(ts.URL+"/value/", "application/json", bytes.NewBufferString(jsonRequest))
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to lcose response body: %s", err)
		}
	}()

	postValueResp, _ := io.ReadAll(resp.Body)
	fmt.Println("POST /value/:", resp.Status)
	fmt.Println("Response:", string(postValueResp))

	// Output:
	// POST /update/: 200 OK
	// GET /value/gauge/Alloc: 123.4
	// POST /value/: 200 OK
	// Response: {"id":"Alloc","type":"gauge","value":123.4}
}
