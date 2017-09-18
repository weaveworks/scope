package middleware

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/weaveworks/common/logging"
)

const gRPC = "gRPC"

// ServerLoggingInterceptor logs gRPC requests, errors and latency.
var ServerLoggingInterceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	begin := time.Now()
	resp, err := handler(ctx, req)
	entry := logging.With(ctx).WithFields(log.Fields{"method": info.FullMethod, "duration": time.Since(begin)})
	if err != nil {
		entry.WithError(err).Warn(gRPC)
	} else {
		entry.Debugf("%s (success)", gRPC)
	}
	return resp, err
}
