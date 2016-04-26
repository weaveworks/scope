package middleware

import (
	"net/http"
	"regexp"
)

// PathRewrite supports regex matching and replace on Request URIs
func PathRewrite(regexp *regexp.Regexp, replacement string) Interface {
	return pathRewrite{
		regexp:      regexp,
		replacement: replacement,
	}
}

type pathRewrite struct {
	regexp      *regexp.Regexp
	replacement string
}

func (p pathRewrite) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.RequestURI = p.regexp.ReplaceAllString(r.RequestURI, p.replacement)
		next.ServeHTTP(w, r)
	})
}
