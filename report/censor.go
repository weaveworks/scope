package report

import (
	"net/http"
	"strings"
)

type keyMatcher func(string) bool

func keyEquals(fixedKey string) keyMatcher {
	return func(key string) bool {
		return key == fixedKey
	}
}

func keyStartsWith(prefix string) keyMatcher {
	return func(key string) bool {
		return strings.HasPrefix(key, prefix)
	}
}

type censorValueFunc func(string) string

// TODO: Implement this in a more systematic way.
func censorTopology(t *Topology, match keyMatcher, censor censorValueFunc) {
	for nodeID := range t.Nodes {
		for entryID := range t.Nodes[nodeID].Latest {
			entry := &t.Nodes[nodeID].Latest[entryID]
			if match(entry.key) {
				entry.Value = censor(entry.Value)
			}
		}
	}
}

// CensorConfig describes how probe reports should
// be censored when rendered through the API.
type CensorConfig struct {
	HideCommandLineArguments bool
	HideEnvironmentVariables bool
}

// GetCensorConfigFromQueryParams extracts censor config from request query params.
func GetCensorConfigFromQueryParams(req *http.Request) CensorConfig {
	return CensorConfig{
		HideCommandLineArguments: true || req.URL.Query().Get("hideCommandLineArguments") == "true",
		HideEnvironmentVariables: true || req.URL.Query().Get("hideEnvironmentVariables") == "true",
	}
}

// CensorRawReport removes any sensitive data from
// the raw report based on the request query params.
func CensorRawReport(r Report, cfg CensorConfig) Report {
	var (
		makeEmpty = func(string) string { return "" }
	)
	if cfg.HideCommandLineArguments {
		censorTopology(&r.Process, keyEquals(Cmdline), StripCommandArgs)
		censorTopology(&r.Container, keyEquals(DockerContainerCommand), StripCommandArgs)
	}
	if cfg.HideEnvironmentVariables {
		censorTopology(&r.Container, keyStartsWith(DockerEnvPrefix), makeEmpty)
	}
	return r
}

// StripCommandArgs removes all the arguments from the command
func StripCommandArgs(command string) string {
	return strings.Split(command, " ")[0]
}
