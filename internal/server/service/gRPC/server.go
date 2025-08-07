// Package servergrpc provides gRPC request handlers and interceptors,
// including logging, hash signature verification, and IP subnet validation.
package servergrpc

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/a-palonskaa/metrics-server/gen/proto"
	metrics "github.com/a-palonskaa/metrics-server/internal/models/metrics"
	usecase "github.com/a-palonskaa/metrics-server/internal/server/usecase"
)

// Params contains the dependencies required to create a ServerHandler.
type Params struct {
	// MsUsecase provides in-memory storage operations for metrics.
	MsUsecase usecase.MemStorage

	// PingUsecase provides health check functionality.
	PingUsecase usecase.Ping

	// Key is a secret used for signature verification or authentication.
	Key string

	// TrustedSubnet is the CIDR subnet string used to validate client IPs.
	TrustedSubnet string
}

// ServerHandler implements the MetricsServiceServer interface and handles
// incoming gRPC requests using the provided business logic.
type ServerHandler struct {
	proto.UnimplementedMetricsServiceServer

	msUsecase     usecase.MemStorage
	pingUsecase   usecase.Ping
	key           string
	trustedSubnet string
}

// NewServerHandler creates a new instance of ServerHandler with the given parameters.
func NewServerHandler(params Params) *ServerHandler {
	return &ServerHandler{
		msUsecase:     params.MsUsecase,
		pingUsecase:   params.PingUsecase,
		key:           params.Key,
		trustedSubnet: params.TrustedSubnet,
	}
}

// CheckConnection handles a health check request to verify backend connectivity.
func (s *ServerHandler) CheckConnection(ctx context.Context, req *proto.CheckConnectionRequest) (*proto.CheckConnectionResponse, error) {
	err := s.pingUsecase.Ping(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "no connection:%v", err)
	}
	return &proto.CheckConnectionResponse{}, nil
}

// UpdateMetric handles a request to update a single metric.
func (s *ServerHandler) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	metric := req.GetMetric()
	m := metrics.ProtoToModel(metric)

	if err := s.msUsecase.UpdateMetrics(ctx, metrics.Metrics{m}); err != nil {
		log.Error().Err(err).Msg("failed to add metric to storage")
		return nil, status.Errorf(codes.InvalidArgument, "failed to add metric to storage:%v", err)
	}
	return &proto.UpdateMetricResponse{Metric: metrics.ModelToProto(m)}, nil
}

// UpdateMetrics handles a request to update a batch of metrics.
func (s *ServerHandler) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	mtr := req.GetMetrics()
	m := metrics.ProtoListToModel(mtr)

	if err := s.msUsecase.UpdateMetrics(ctx, m); err != nil {
		log.Error().Err(err).Msg("failed to add metrics to storage")
		return nil, status.Errorf(codes.InvalidArgument, "failed to add metrics to storage:%v", err)
	}
	return &proto.UpdateMetricsResponse{Metrics: metrics.ModelListToProto(m)}, nil
}

// ListAllMetrics handles a request to retrieve all available metrics.
func (s *ServerHandler) ListAllMetrics(ctx context.Context, _ *proto.ListAllMetricsRequest) (*proto.ListAllMetricsResponse, error) {
	return &proto.ListAllMetricsResponse{
		Metrics: metrics.ModelListToProto(metrics.Metrics(s.msUsecase.GetAllMetrics(ctx))),
	}, nil
}

// GetMetric handles a request to retrieve a single metric by name and type.
func (s *ServerHandler) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	metric := req.GetMetric()
	m := metrics.ProtoToModel(metric)

	if err := s.msUsecase.GetMetric(ctx, &m); err != nil {
		log.Info().Err(err).Msg("failed to get metric")
		return nil, status.Errorf(codes.InvalidArgument, "failed to get metric:%v", err)
	}
	return &proto.GetMetricResponse{Metric: metrics.ModelToProto(m)}, nil
}
