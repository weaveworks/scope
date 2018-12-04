package multitenant

import (
	"fmt"
	"net/http"

	"context"

	"github.com/weaveworks/scope/app"
)

// ErrUserIDNotFound should be returned by a UserIDer when it fails to ID the
// user for a request.
var ErrUserIDNotFound = fmt.Errorf("User ID not found")

// UserIDer identifies users given a request context.
type UserIDer func(context.Context) (string, error)

// UserIDHeader returns a UserIDer which a header by the supplied key.
func UserIDHeader(headerName string) UserIDer {
	return func(ctx context.Context) (string, error) {
		request, ok := ctx.Value(app.RequestCtxKey).(*http.Request)
		if !ok || request == nil {
			return "", ErrUserIDNotFound
		}
		userID := request.Header.Get(headerName)
		if userID == "" {
			return "", ErrUserIDNotFound
		}
		return userID, nil
	}
}

// NoopUserIDer always returns the empty user ID.
func NoopUserIDer(context.Context) (string, error) {
	return "", nil
}
