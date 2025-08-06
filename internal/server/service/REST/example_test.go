package serverrest_test

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
	server "github.com/a-palonskaa/metrics-server/internal/server/service/REST"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

func setUp() *chi.Mux {
	r := chi.NewRouter()

	msUsecase := usecase.NewMemStorage(memstorage.New())
	pingUsecase := usecase.NewPing(database.NewConn(""))

	serverHandler := server.New(server.Params{
		MsUsecase:   msUsecase,
		PingUsecase: pingUsecase,
	})

	_ = serverHandler.Router(r)
	return r
}

// Example_postUpdateMetric demonstrates sending and retrieving a metric
// using the HTTP endpoints of the metrics server.
func Example_updateMetric() {
	// Launch test HTTP server
	ts := httptest.NewServer(setUp())
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

	// Output:
	// POST /update/: 200 OK
}

// Example_getMetricByURL demonstrates getting a metric by type and name
func Example_getMetricByURL() {
	// Launch test HTTP server
	ts := httptest.NewServer(setUp())
	defer ts.Close()

	// Get metric via GET /value/{type}/{name}
	resp, err := http.Get(ts.URL + "/value/gauge/Alloc")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close response body: %s", err)
		}
	}()

	getRespBody, _ := io.ReadAll(resp.Body)
	fmt.Println("GET /value/gauge/Alloc:", string(getRespBody))

	// Output:
	// GET /value/gauge/Alloc: Metric not found
}
