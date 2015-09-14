package sanitize

import (
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
)

// URL returns a function that sanitizes a URL string. It lets underspecified
// strings to be converted to usable URLs via some default arguments.
func URL(scheme string, port int, path string) func(string) string {
	if scheme == "" {
		scheme = "http://"
	}
	return func(s string) string {
		if s == "" {
			return s // can't do much here
		}
		if !strings.HasPrefix(s, "http") {
			s = scheme + s
		}
		u, err := url.Parse(s)
		if err != nil {
			log.Printf("%q: %v", s, err)
			return s // oh well
		}
		if port > 0 {
			if _, _, err = net.SplitHostPort(u.Host); err != nil {
				u.Host += fmt.Sprintf(":%d", port)
			}
		}
		if path != "" && u.Path != path {
			u.Path = path
		}
		return u.String()
	}
}
