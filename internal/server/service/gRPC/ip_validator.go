package servergrpc

import (
	"context"
	"net"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// IPValidationInterceptor returns a gRPC interceptor that validates the client's IP address against a trusted subnet.
func IPValidationInterceptor(trustedSubnet string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if trustedSubnet != "" {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Error(codes.Internal, "no metadata in context")
			}
			ips := md.Get("x-real-ip")
			if len(ips) == 0 {
				return nil, status.Error(codes.Unauthenticated, "x-real-ip header required")
			}
			ip := ips[0]

			if code := isIPtrustful(ip, trustedSubnet); code != codes.OK {
				return nil, status.Error(code, "not trustful subnet")
			}
		}
		return handler(ctx, req)
	}
}

func isIPtrustful(ipRaw string, trustedSubnet string) codes.Code {
	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		log.Info().Err(err).Msg("invalid trusted subnet")
		return codes.Internal
	}

	ip := net.ParseIP(ipRaw)
	if ip == nil {
		log.Info().Err(err).Msgf("ip %s is not a valid textual representation of an IP address", ipRaw)
		return codes.DataLoss
	}

	if !subnet.Contains(ip) {
		return codes.PermissionDenied
	}
	return codes.OK
}
