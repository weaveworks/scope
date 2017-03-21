package user

import (
	"net/http"

	"golang.org/x/net/context"
)

// ExtractFromHTTPRequest extracts the user ID from the request headers and returns
// the user ID and a context with the user ID embbedded.
func ExtractFromHTTPRequest(r *http.Request) (string, context.Context, error) {
	userID := r.Header.Get(orgIDHeaderName)
	if userID == "" {
		return "", r.Context(), ErrNoUserID
	}
	return userID, Inject(r.Context(), userID), nil
}

// InjectIntoHTTPRequest injects the userID from the context into the request headers.
func InjectIntoHTTPRequest(ctx context.Context, r *http.Request) error {
	userID, err := Extract(ctx)
	if err != nil {
		return err
	}
	existingID := r.Header.Get(orgIDHeaderName)
	if existingID != "" && existingID != userID {
		return ErrDifferentIDPresent
	}
	r.Header.Set(orgIDHeaderName, userID)
	return nil
}
