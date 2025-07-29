// Package agent provides functionality for collecting and sending runtime and system
// metrics from a agent to a server.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"sync"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	usecase "github.com/a-palonskaa/metrics-server/internal/agent/usecase"
	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	errhandler "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
	hash "github.com/a-palonskaa/metrics-server/pkg/hash"
)

// AgentHadler defines a handler wrapping MemStorage usecase
type AgentHandler struct {
	msUsecase usecase.MemStorage
}

// NewHandler creates a new AgentHandler instance using the provided metrics repository.
func NewHandler(storage usecase.MetricsRepository) AgentHandler {
	return AgentHandler{
		msUsecase: usecase.NewMemStorageUsecase(storage),
	}
}

// SendMetrics sends all stored runtime metrics to the metrics server.
// The request is retried if it fails, using a retriable error handler.
func (h AgentHandler) SendMetrics(ctx context.Context, client *resty.Client, key string) error {
	body := h.msUsecase.ListAllMetrics(ctx)

	err := errhandler.RetriableErrHadlerVoid(
		func() error {
			return h.SendRequest(client, body, key)
		}, errhandler.CompareErrAgent)
	if err != nil {
		log.Error().Err(err).Msg("error sending metrics")
		return fmt.Errorf("error sending metrics: %v", err)
	}
	log.Info().Msg("send metrics successfully")
	return nil
}

// UpdateRuntimeMetrics updates runtime metrics such as memory stats, GC, etc.
func (h AgentHandler) UpdateRuntimeMetrics(ctx context.Context) {
	if err := h.msUsecase.UpdateMetrics(ctx); err != nil {
		log.Error().Err(err).Msg("failed to update metrics")
	}
	log.Info().Msg("send metrics successfully")
}

// UpdateSystemMetrics updates system metrics such as CPU load and disk usage.
func (h AgentHandler) UpdateSystemMetrics(ctx context.Context) {
	if err := h.msUsecase.UpdateSysMetrics(ctx); err != nil {
		log.Error().Err(err).Msg("failed to update metrics")
	}
}

// gzipWriterPool is a sync.Pool that reuses gzip.Writer instances
// to minimize allocation overhead during metrics compression.
var gzipWriterPool = sync.Pool{
	New: func() any {
		gz, err := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		if err != nil {
			log.Error().Err(err).Msg("failed to create gzip writer")
		}
		return gz
	},
}

// SendRequest prepares the metrics list, compresses it using gzip,
// optionally signs it with SHA256 hash, and sends it to the metrics server
// via a POST request to the /updates/ endpoint.
func (h AgentHandler) SendRequest(client *resty.Client, body metrics.Metrics, key string) error {
	if len(body) == 0 {
		return nil
	}

	mtr := body.Deserialize()
	jsonData, err := mtr.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %v", err)
	}

	var buf bytes.Buffer
	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		log.Error().Err(err).Msg("failed to write compressed data")
		gzipWriterPool.Put(gz)
		return fmt.Errorf("failed to write compressed data: %v", err)
	}
	if err := gz.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close gzip writer")
		gzipWriterPool.Put(gz)
		return fmt.Errorf("failed to close gzip writer: %v", err)
	}
	gzipWriterPool.Put(gz)

	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes())

	if key != "" {
		dst, err := hash.Calculate([]byte(key), jsonData)
		if err != nil {
			log.Error().Err(err).Msg("failed to calculate hash")
		}
		req.SetHeader("HashSHA256", hex.EncodeToString(dst))
	}

	_, err = req.Post("/updates/")
	if err != nil {
		log.Error().Err(err).Msgf("failed to send request w metrics update to server")
		return fmt.Errorf("failed to send request w metrics update to server: %v", err)
	}
	return nil
}
