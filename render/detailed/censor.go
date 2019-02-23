package detailed

import (
	"github.com/weaveworks/scope/report"
)

func censorNodeSummary(s NodeSummary, cfg report.CensorConfig) NodeSummary {
	if cfg.HideCommandLineArguments && s.Metadata != nil {
		// Iterate through all the metadata rows and strip the
		// arguments from all the values containing a command
		// (while making sure everything is done in a non-mutable way).
		metadata := []report.MetadataRow{}
		for _, row := range s.Metadata {
			if report.IsCommandEntry(row.ID) {
				row.Value = report.StripCommandArgs(row.Value)
			}
			metadata = append(metadata, row)
		}
		s.Metadata = metadata
	}
	if cfg.HideEnvironmentVariables && s.Tables != nil {
		// Copy across all the tables except the environment
		// variable ones (ensuring the operation is non-mutable).
		tables := []report.Table{}
		for _, table := range s.Tables {
			if !report.IsEnvironmentVarsEntry(table.ID) {
				tables = append(tables, table)
			}
		}
		s.Tables = tables
	}
	return s
}

// CensorNode removes any sensitive data from a node.
func CensorNode(node Node, cfg report.CensorConfig) Node {
	node.NodeSummary = censorNodeSummary(node.NodeSummary, cfg)
	return node
}

// CensorNodeSummaries removes any sensitive data from a list of node summaries.
func CensorNodeSummaries(summaries NodeSummaries, cfg report.CensorConfig) NodeSummaries {
	censored := NodeSummaries{}
	for key := range summaries {
		censored[key] = censorNodeSummary(summaries[key], cfg)
	}
	return censored
}
