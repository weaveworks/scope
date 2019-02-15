package report

import "strings"

// import log "github.com/sirupsen/logrus"

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

func assignEmpty(key string) string {
	return ""
}

func censorTopology(t *Topology, match keyMatcher, censor censorValueFunc) {
	for nodeID := range t.Nodes {
		for entryID := range t.Nodes[nodeID].Latest {
			entry := &t.Nodes[nodeID].Latest[entryID]
			if match(entry.key) {
				// log.Infof("Blabla ... %s ... %s ... %s", entry.key, entry.Value, censor(entry.Value))
				entry.Value = censor(entry.Value)
			}
		}
	}
}

// CensorConfig describe which parts of the report needs to be censored.
type CensorConfig struct {
	HideCommandLineArguments bool
	HideEnvironmentVariables bool
}

// CensorReport removes any sensitive data from the report.
func CensorReport(r Report, cfg CensorConfig) Report {
	if cfg.HideCommandLineArguments {
		censorTopology(&r.Process, keyEquals(Cmdline), StripCommandArgs)
		censorTopology(&r.Container, keyEquals(DockerContainerCommand), StripCommandArgs)
	}
	if cfg.HideEnvironmentVariables {
		censorTopology(&r.Container, keyStartsWith(DockerEnvPrefix), assignEmpty)
	}
	return r
}
