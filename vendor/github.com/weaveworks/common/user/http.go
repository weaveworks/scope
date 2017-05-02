package user

import (
	"net/http"

	"golang.org/x/net/context"
)

const (
	// orgIDHeaderName is a legacy from scope as a service.
	orgIDHeaderName  = "X-Scope-OrgID"
	userIDHeaderName = "X-Scope-UserID"

	// LowerOrgIDHeaderName as gRPC / HTTP2.0 headers are lowercased.
	lowerOrgIDHeaderName = "x-scope-orgid"
)

// ExtractOrgIDFromHTTPRequest extracts the org ID from the request headers and returns
// the org ID and a context with the org ID embedded.
func ExtractOrgIDFromHTTPRequest(r *http.Request) (string, context.Context, error) {
	orgID := r.Header.Get(orgIDHeaderName)
	if orgID == "" {
		return "", r.Context(), ErrNoOrgID
	}
	return orgID, InjectOrgID(r.Context(), orgID), nil
}

// InjectOrgIDIntoHTTPRequest injects the orgID from the context into the request headers.
func InjectOrgIDIntoHTTPRequest(ctx context.Context, r *http.Request) error {
	orgID, err := ExtractOrgID(ctx)
	if err != nil {
		return err
	}
	existingID := r.Header.Get(orgIDHeaderName)
	if existingID != "" && existingID != orgID {
		return ErrDifferentOrgIDPresent
	}
	r.Header.Set(orgIDHeaderName, orgID)
	return nil
}

// ExtractUserIDFromHTTPRequest extracts the org ID from the request headers and returns
// the org ID and a context with the org ID embedded.
func ExtractUserIDFromHTTPRequest(r *http.Request) (string, context.Context, error) {
	userID := r.Header.Get(userIDHeaderName)
	if userID == "" {
		return "", r.Context(), ErrNoUserID
	}
	return userID, InjectUserID(r.Context(), userID), nil
}

// InjectUserIDIntoHTTPRequest injects the userID from the context into the request headers.
func InjectUserIDIntoHTTPRequest(ctx context.Context, r *http.Request) error {
	userID, err := ExtractUserID(ctx)
	if err != nil {
		return err
	}
	existingID := r.Header.Get(userIDHeaderName)
	if existingID != "" && existingID != userID {
		return ErrDifferentUserIDPresent
	}
	r.Header.Set(userIDHeaderName, userID)
	return nil
}
