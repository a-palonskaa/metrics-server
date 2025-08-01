package serverGRPC

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func LoggerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	reqSize := 0
	if msg, ok := req.(proto.Message); ok {
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Info().Err(err).Msg("failed to marshal request for size logging")
		} else {
			reqSize = len(data)
		}
	}

	resp, err := handler(ctx, req)

	statusCode := codes.OK
	if err != nil {
		if st, ok := status.FromError(err); ok {
			statusCode = st.Code()
		} else {
			statusCode = codes.Unknown
		}
	}

	log.Info().Str("method", info.FullMethod).Int("request_size", reqSize).Str("status_code", statusCode.String()).Msg("request-response")
	return resp, err
}
