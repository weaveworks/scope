package middleware

import (
	"net/http"
	"net/url"
)

// Redirect middleware, will redirect requests to hosts which match any of the
// Matches to RedirectScheme://RedirectHost
type Redirect struct {
	Matches []Match

	RedirectHost   string
	RedirectScheme string
}

// Match specifies a match for a redirect.  Host and/or Scheme can be empty
// signify match-all.
type Match struct {
	Host, Scheme string
}

func (m Match) match(u *url.URL) bool {
	if m.Host != "" && m.Host != u.Host {
		return false
	}

	if m.Scheme != "" && m.Scheme != u.Scheme {
		return false
	}

	return true
}

// Wrap implements Middleware
func (m Redirect) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, match := range m.Matches {
			if match.match(r.URL) {
				r.URL.Host = m.RedirectHost
				r.URL.Scheme = m.RedirectScheme
				http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
