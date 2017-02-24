package user

import (
	"fmt"

	"golang.org/x/net/context"
)

// UserIDContextKey is the key used in contexts to find the userid
type contextKey int

const userIDContextKey contextKey = 0

// OrgIDHeaderName is a legacy from scope as a service.
const OrgIDHeaderName = "X-Scope-OrgID"

// LowerOrgIDHeaderName as gRPC / HTTP2.0 headers are lowercased.
const LowerOrgIDHeaderName = "x-scope-orgid"

// GetID returns the user
func GetID(ctx context.Context) (string, error) {
	userid, ok := ctx.Value(userIDContextKey).(string)
	if !ok {
		return "", fmt.Errorf("no user id")
	}
	return userid, nil
}

// WithID returns a derived context containing the user ID.
func WithID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, interface{}(userIDContextKey), userID)
}
