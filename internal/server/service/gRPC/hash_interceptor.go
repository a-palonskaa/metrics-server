package servergrpc

import (
	"context"
	"encoding/hex"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	hash "github.com/a-palonskaa/metrics-server/pkg/hash"
)

// HashInterceptor returns a gRPC unary interceptor that verifies the request hash and signs the response.
func HashInterceptor(key string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if key == "" {
			return handler(ctx, req)
		}

		msg, ok := req.(proto.Message)
		if !ok {
			log.Info().Msg("request is not a proto.Message")
			return handler(ctx, req)
		}
		data, err := proto.Marshal(msg)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal request for hash verification")
			return nil, status.Errorf(codes.InvalidArgument, "invalid request format:%v", err)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			if hashStrs := md.Get("hashsha256"); len(hashStrs) > 0 {
				hashStr := hashStrs[0]
				expectedHash, err := hex.DecodeString(hashStr)
				if err != nil {
					log.Error().Err(err).Msg("failed to decode hash header")
					return nil, status.Errorf(codes.InvalidArgument, "invalid hash format")
				}

				if !hash.Verify([]byte(key), data, expectedHash) {
					log.Info().Msg("hash mismatch")
					return nil, status.Errorf(codes.NotFound, "hash mismatch")
				}
			}
		}

		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		msg, ok = resp.(proto.Message)
		if !ok {
			log.Warn().Msg("response is not a proto.Message")
			return resp, nil
		}
		data, err = proto.Marshal(msg)
		if err != nil {
			log.Error().Err(err).Msg("failed to marshal response for hash signing")
			return resp, nil
		}

		signedHash, err := hash.Calculate([]byte(key), data)
		if err != nil {
			log.Error().Err(err).Msg("failed to calculate response hash")
			return resp, nil
		}

		if err := grpc.SetHeader(ctx, metadata.Pairs("hashsha256", hex.EncodeToString(signedHash))); err != nil {
			log.Info().Err(err).Msg("failed to set hashsha256 header")
		}
		return resp, nil
	}
}
