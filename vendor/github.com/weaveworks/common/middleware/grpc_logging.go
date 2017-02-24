package middleware

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const gRPC = "gRPC"

// ServerLoggingInterceptor logs gRPC requests, errors and latency.
func ServerLoggingInterceptor(logSuccess bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		begin := time.Now()
		resp, err := handler(ctx, req)
		if err != nil {
			log.Errorf("%s %s (%v) %s", gRPC, info.FullMethod, err, time.Since(begin))
		} else if logSuccess {
			log.Infof("%s %s (success) %s", gRPC, info.FullMethod, time.Since(begin))
		}
		return resp, err
	}
}
