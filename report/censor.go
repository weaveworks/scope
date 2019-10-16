package report

import (
	"net/http"
	"strings"
)

// CensorConfig describes how probe reports should
// be censored when rendered through the API.
type CensorConfig struct {
	HideCommandLineArguments bool
	HideEnvironmentVariables bool
}

// GetCensorConfigFromRequest extracts censor config from request query params.
func GetCensorConfigFromRequest(req *http.Request) CensorConfig {
	return CensorConfig{
		HideCommandLineArguments: req.URL.Query().Get("hideCommandLineArguments") == "true",
		HideEnvironmentVariables: req.URL.Query().Get("hideEnvironmentVariables") == "true",
	}
}

// IsCommandEntry returns true iff the entry comes from a command line
// that might need to be conditionally censored.
func IsCommandEntry(key string) bool {
	return key == Cmdline || key == DockerContainerCommand
}

// IsEnvironmentVarsEntry returns true if the entry might expose some
// environment variables data might need to be conditionally censored.
func IsEnvironmentVarsEntry(key string) bool {
	return strings.HasPrefix(key, DockerEnvPrefix)
}

// StripCommandArgs removes all the arguments from the command
func StripCommandArgs(command string) string {
	return strings.Split(command, " ")[0]
}

// CensorRawReport removes any sensitive data from the raw report based on the request query params.
func CensorRawReport(rawReport Report, cfg CensorConfig) Report {
	// Create a copy of the report first to make sure the operation is immutable.
	censoredReport := rawReport.Copy()
	censoredReport.ID = rawReport.ID

	censoredReport.WalkTopologies(func(t *Topology) {
		for nodeID, node := range t.Nodes {
			if node.Latest != nil {
				latest := make([]stringLatestEntry, 0, len(node.Latest))
				for _, entry := range node.Latest {
					// If environment variables are to be hidden, omit passing them to the final report.
					if cfg.HideEnvironmentVariables && IsEnvironmentVarsEntry(entry.key) {
						continue
					}
					// If command line arguments are to be hidden, strip them away.
					if cfg.HideCommandLineArguments && IsCommandEntry(entry.key) {
						entry.value = StripCommandArgs(entry.value)
					}
					// Pass the latest entry to the final report.
					latest = append(latest, entry)
				}
				node.Latest = latest
				t.Nodes[nodeID] = node
			}
		}
	})
	return censoredReport
}
