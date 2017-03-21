package user

import (
	"golang.org/x/net/context"

	"github.com/weaveworks/common/errors"
)

type contextKey int

const (
	// UserIDContextKey is the key used in contexts to find the userid
	userIDContextKey contextKey = 0

	// orgIDHeaderName is a legacy from scope as a service.
	orgIDHeaderName = "X-Scope-OrgID"

	// LowerOrgIDHeaderName as gRPC / HTTP2.0 headers are lowercased.
	lowerOrgIDHeaderName = "x-scope-orgid"
)

// Errors that we return
const (
	ErrNoUserID           = errors.Error("no user id")
	ErrDifferentIDPresent = errors.Error("different user ID already present")
	ErrTooManyUserIDs     = errors.Error("multiple user IDs present")
)

// Extract gets the user ID from the context
func Extract(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	if !ok {
		return "", ErrNoUserID
	}
	return userID, nil
}

// Inject returns a derived context containing the user ID.
func Inject(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, interface{}(userIDContextKey), userID)
}
