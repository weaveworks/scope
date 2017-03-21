package user

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// ExtractFromGRPCRequest extracts the user ID from the request metadata and returns
// the user ID and a context with the user ID injected.
func ExtractFromGRPCRequest(ctx context.Context) (string, context.Context, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", ctx, ErrNoUserID
	}

	userIDs, ok := md[lowerOrgIDHeaderName]
	if !ok || len(userIDs) != 1 {
		return "", ctx, ErrNoUserID
	}

	return userIDs[0], Inject(ctx, userIDs[0]), nil
}

// InjectIntoGRPCRequest injects the userID from the context into the request metadata.
func InjectIntoGRPCRequest(ctx context.Context) (context.Context, error) {
	userID, err := Extract(ctx)
	if err != nil {
		return ctx, err
	}

	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	newCtx := ctx
	if userIDs, ok := md[lowerOrgIDHeaderName]; ok {
		if len(userIDs) == 1 {
			if userIDs[0] != userID {
				return ctx, ErrDifferentIDPresent
			}
		} else {
			return ctx, ErrTooManyUserIDs
		}
	} else {
		md = md.Copy()
		md[lowerOrgIDHeaderName] = []string{userID}
		newCtx = metadata.NewContext(ctx, md)
	}

	return newCtx, nil
}
