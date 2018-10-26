package main

import (
	"bytes"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

// BasicAuthentication middleware authenticate http request
type BasicAuthentication struct {
	Realm    string
	User     string
	Password string
}

// Wrap implements Middleware
func (b BasicAuthentication) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authenticated := b.authenticate(r)
		if !authenticated {
			b.requestAuth(w, r)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
func (b *BasicAuthentication) authenticate(r *http.Request) bool {
	const basicScheme string = "Basic "
	// Confirm the request is sending Basic Authentication credentials.
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, basicScheme) {
		return false
	}
	str, err := base64.StdEncoding.DecodeString(auth[len(basicScheme):])
	if err != nil {
		return false
	}

	creds := bytes.SplitN(str, []byte(":"), 2)

	if len(creds) != 2 {
		return false
	}

	givenUser := sha256.Sum256([]byte(string(creds[0])))
	givenPass := sha256.Sum256([]byte(string(creds[1])))
	requiredUser := sha256.Sum256([]byte(b.User))
	requiredPass := sha256.Sum256([]byte(b.Password))
	// Compare the supplied credentials to those set in our options
	if subtle.ConstantTimeCompare(givenUser[:], requiredUser[:]) == 1 &&
		subtle.ConstantTimeCompare(givenPass[:], requiredPass[:]) == 1 {
		return true
	}

	return false
}

func (b *BasicAuthentication) requestAuth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Basic realm=%q`, b.Realm))
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
	return
}
