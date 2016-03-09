package multitenant

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/weaveworks/scope/app"
)

// ErrNotFound should be returned by a UserIDer when it fails to ID the
// user for a request.
var ErrNotFound = fmt.Errorf("User ID not found")

// UserIDer identifies users given a request context.
type UserIDer func(context.Context) (string, error)

// UserIDHeader returns a UserIDer which a header by the supplied key.
func UserIDHeader(headerName string) UserIDer {
	return func(ctx context.Context) (string, error) {
		request, ok := ctx.Value(app.RequestCtxKey).(*http.Request)
		if !ok || request == nil {
			return "", ErrNotFound
		}
		userID := request.Header.Get(headerName)
		if userID == "" {
			return "", ErrNotFound
		}
		return userID, nil
	}
}

// NoopUserIDer always returns the empty user ID.
func NoopUserIDer(context.Context) (string, error) {
	return "", nil
}
