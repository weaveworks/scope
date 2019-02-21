package detailed

import (
	"github.com/weaveworks/scope/report"
)

func censorNodeSummary(s *NodeSummary, cfg report.CensorConfig) {
	if cfg.HideCommandLineArguments {
		// Iterate through all the metadata rows and strip the
		// arguments from all the values containing a command.
		for index := range s.Metadata {
			row := &s.Metadata[index]
			if report.IsCommandEntry(row.ID) {
				row.Value = report.StripCommandArgs(row.Value)
			}
		}
	}
	if cfg.HideEnvironmentVariables {
		// Go through all the tables and if environment variables
		// table is found, drop it from the list and stop the loop.
		for index, table := range s.Tables {
			if report.IsEnvironmentVarsEntry(table.ID) {
				s.Tables = append(s.Tables[:index], s.Tables[index+1:]...)
				break
			}
		}
	}
}

// CensorNode removes any sensitive data from a node.
func CensorNode(n Node, cfg report.CensorConfig) Node {
	censorNodeSummary(&n.NodeSummary, cfg)
	return n
}

// CensorNodeSummaries removes any sensitive data from a list of node summaries.
func CensorNodeSummaries(ns NodeSummaries, cfg report.CensorConfig) NodeSummaries {
	for key, summary := range ns {
		censorNodeSummary(&summary, cfg)
		ns[key] = summary
	}
	return ns
}
