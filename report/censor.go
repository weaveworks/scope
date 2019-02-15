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

// CensorReportForRequest removes any sensitive data
// from the report based on the request query params.
func CensorReportForRequest(rep Report, req *http.Request) Report {
	var (
		hideCommandLineArguments = req.URL.Query().Get("hideCommandLineArguments") == "true"
		hideEnvironmentVariables = req.URL.Query().Get("hideEnvironmentVariables") == "true"
		makeEmpty                = func(string) string { return "" }
	)
	if hideCommandLineArguments {
		censorTopology(&rep.Process, keyEquals(Cmdline), StripCommandArgs)
		censorTopology(&rep.Container, keyEquals(DockerContainerCommand), StripCommandArgs)
	}
	if hideEnvironmentVariables {
		censorTopology(&rep.Container, keyStartsWith(DockerEnvPrefix), makeEmpty)
	}
	return rep
}

// StripCommandArgs removes all the arguments from the command
func StripCommandArgs(command string) string {
	return strings.Split(command, " ")[0]
}
