package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"

	usecase "github.com/a-palonskaa/metrics-server/internal/agent/usecase"
	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	errhandler "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
)

type AgentHandler struct {
	msUsecase usecase.MemStorage
}

func NewHandler(storage usecase.MetricsRepository) AgentHandler {
	return AgentHandler{
		msUsecase: usecase.NewMemStorageUsecase(storage),
	}
}

func (h AgentHandler) SendMetrics(ctx context.Context, client *resty.Client, endpointAddr string) {
	err := errhandler.RetriableErrHadlerVoid(
		func() error {
			body := h.msUsecase.ListAllMetrics(ctx)
			return h.SendRequest(client, endpointAddr, body)
		}, errhandler.CompareErrAgent)
	if err != nil {
		log.Error().Err(err).Msg("error sending metrics")
	}
}

func (h AgentHandler) Update(ctx context.Context) {
	if err := h.msUsecase.UpdateMetrics(ctx); err != nil {
		log.Error().Err(err).Msg("failed to update metrics")
	}
}

func (h AgentHandler) SendRequest(client *resty.Client, endpoint string, body metrics.Metrics) error {
	if len(body) == 0 {
		return nil
	}

	mtr := body.Deserialize()
	jsonData, err := mtr.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %v", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		log.Error().Err(err).Msg("failed to write compressed data")
		return fmt.Errorf("failed to write compressed data: %v", err)
	}
	if err := gz.Close(); err != nil {
		log.Error().Err(err).Msg("failed to close gzip Writer")
		return fmt.Errorf("failed to close gzip Writer: %v", err)
	}

	_, err = client.SetBaseURL("http://"+endpoint).R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post("/updates/")
	if err != nil {
		log.Error().Err(err).Msgf("failed to send request w metrics update to server %s", endpoint)
		return fmt.Errorf("failed to send request w metrics update to server: %v", err)
	}
	return nil
}
