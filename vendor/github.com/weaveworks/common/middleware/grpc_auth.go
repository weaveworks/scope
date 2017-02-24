package middleware

import (
	"fmt"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/weaveworks/common/user"
)

// ClientUserHeaderInterceptor propagates the user ID from the context to gRPC metadata, which eventually ends up as a HTTP2 header.
func ClientUserHeaderInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	userID, err := user.GetID(ctx)
	if err != nil {
		return err
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}

	newCtx := ctx
	if userIDs, ok := md[user.LowerOrgIDHeaderName]; ok {
		switch len(userIDs) {
		case 1:
			if userIDs[0] != userID {
				return fmt.Errorf("wrong user ID found")
			}
		default:
			return fmt.Errorf("multiple user IDs found")
		}
	} else {
		md = md.Copy()
		md[user.LowerOrgIDHeaderName] = []string{userID}
		newCtx = metadata.NewContext(ctx, md)
	}

	return invoker(newCtx, method, req, reply, cc, opts...)
}

// ServerUserHeaderInterceptor propagates the user ID from the gRPC metadata back to our context.
func ServerUserHeaderInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("no metadata")
	}

	userIDs, ok := md[user.LowerOrgIDHeaderName]
	if !ok || len(userIDs) != 1 {
		return nil, fmt.Errorf("no user id")
	}

	newCtx := user.WithID(ctx, userIDs[0])
	return handler(newCtx, req)
}
