// Package agent provides functionality for collecting and sending runtime and system
// metrics from a agent to a server with gRPC API.
package agentgrpc

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	pt "google.golang.org/protobuf/proto"

	proto "github.com/a-palonskaa/metrics-server/gen/proto"
	usecase "github.com/a-palonskaa/metrics-server/internal/agent/usecase"
	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	errhandler "github.com/a-palonskaa/metrics-server/pkg/err_handlers"
	hash "github.com/a-palonskaa/metrics-server/pkg/hash"
)

// AgentHadler defines a handler wrapping MemStorage usecase
type AgentHandler struct {
	msUsecase usecase.MemStorage
	client    proto.MetricsServiceClient
}

// NewHandler creates a new AgentHandler instance using the provided metrics repository.
func NewHandler(storage usecase.MetricsRepository, client proto.MetricsServiceClient) AgentHandler {
	return AgentHandler{
		msUsecase: usecase.NewMemStorageUsecase(storage),
		client:    client,
	}
}

// HashSigning provides an Interceptor for hash signing of client requests.
func HashSigning(key string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if key != "" {
			if msg, ok := req.(pt.Message); ok {
				data, err := pt.Marshal(msg)
				if err != nil {
					log.Info().Err(err).Msg("failed to marshal request for hash signing")
				} else {
					hashBytes, err := hash.Calculate([]byte(key), data)
					if err != nil {
						log.Info().Err(err).Msg("failed to calculate hash")
					}
					ctx = metadata.AppendToOutgoingContext(ctx, "hashsha256", hex.EncodeToString(hashBytes))
				}
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// SendMetrics sends all stored runtime metrics to the metrics server.
// The request is retried if it fails, using a retriable error handler.
func (h AgentHandler) SendMetrics(ctx context.Context) error {
	body := h.msUsecase.ListAllMetrics(ctx)

	_, err := errhandler.RetriableErrHadler(
		func() (*proto.UpdateMetricsResponse, error) {
			return h.client.UpdateMetrics(ctx, &proto.UpdateMetricsRequest{Metrics: metrics.ModelListToProto(body)})
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
