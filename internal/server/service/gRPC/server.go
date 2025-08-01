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

type Params struct {
	MsUsecase     usecase.MemStorage
	PingUsecase   usecase.Ping
	Key           string
	TrustedSubnet string
}

type ServerHandler struct {
	proto.UnimplementedMetricsServiceServer

	msUsecase     usecase.MemStorage
	pingUsecase   usecase.Ping
	key           string
	trustedSubnet string
}

func NewServerHandler(params Params) *ServerHandler {
	return &ServerHandler{
		msUsecase:     params.MsUsecase,
		pingUsecase:   params.PingUsecase,
		key:           params.Key,
		trustedSubnet: params.TrustedSubnet,
	}
}

func (s *ServerHandler) CheckConnection(ctx context.Context, req *proto.CheckConnectionRequest) (*proto.CheckConnectionResponse, error) {
	err := s.pingUsecase.Ping(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, `no connection: %v`, err)
	}
	return &proto.CheckConnectionResponse{}, nil
}

func (s *ServerHandler) UpdateMetric(ctx context.Context, req *proto.UpdateMetricRequest) (*proto.UpdateMetricResponse, error) {
	metric := req.GetMetric()
	m := metrics.ProtoToModel(metric)

	if err := s.msUsecase.UpdateMetrics(ctx, metrics.Metrics{m}); err != nil {
		log.Error().Err(err).Msg("failed to add metric to storage")
		return nil, status.Errorf(codes.InvalidArgument, `failed to add metric to storage: %v`, err)
	}
	return &proto.UpdateMetricResponse{Metric: metrics.ModelToProto(m)}, nil
}

func (s *ServerHandler) UpdateMetrics(ctx context.Context, req *proto.UpdateMetricsRequest) (*proto.UpdateMetricsResponse, error) {
	mtr := req.GetMetrics()
	m := metrics.ProtoListToModel(mtr)

	if err := s.msUsecase.UpdateMetrics(ctx, m); err != nil {
		log.Error().Err(err).Msg("failed to add metrics to storage")
		return nil, status.Errorf(codes.InvalidArgument, `failed to add metrics to storage: %v`, err)
	}
	return &proto.UpdateMetricsResponse{Metrics: metrics.ModelListToProto(m)}, nil
}

func (s *ServerHandler) ListAllMetrics(ctx context.Context, _ *proto.ListAllMetricsRequest) (*proto.ListAllMetricsResponse, error) {
	return &proto.ListAllMetricsResponse{Metrics: metrics.ModelListToProto(metrics.Metrics(s.msUsecase.GetAllMetrics(ctx)))}, nil
}

func (s *ServerHandler) GetMetric(ctx context.Context, req *proto.GetMetricRequest) (*proto.GetMetricResponse, error) {
	metric := req.GetMetric()
	m := metrics.ProtoToModel(metric)

	if err := s.msUsecase.GetMetric(ctx, &m); err != nil {
		log.Info().Err(err).Msg("failed to get metric")
		return nil, status.Errorf(codes.InvalidArgument, "failed to get metric: %v", err)
	}
	return &proto.GetMetricResponse{Metric: metrics.ModelToProto(m)}, nil
}
