package sanitize

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// URL returns a function that sanitizes a URL string. It lets underspecified
// strings to be converted to usable URLs via some default arguments.
func URL(defaultScheme string, defaultPort int, defaultPath string) func(string) string {
	if defaultScheme == "" {
		defaultScheme = "http://"
	}
	return func(s string) string {
		if s == "" {
			return s // can't do much here
		}
		if !strings.Contains(s, "://") {
			s = defaultScheme + s
		}
		u, err := url.Parse(s)
		if err != nil {
			log.Errorf("%q: %v", s, err)
			return s // oh well
		}
		if _, port, err := net.SplitHostPort(u.Host); err != nil && defaultPort > 0 {
			u.Host += fmt.Sprintf(":%d", defaultPort)
		} else if port == "443" {
			if u.Scheme == "ws" {
				u.Scheme = "wss"
			} else {
				u.Scheme = "https"
			}
		}
		if defaultPath != "" && u.Path != defaultPath {
			u.Path = defaultPath
		}
		return u.String()
	}
}
