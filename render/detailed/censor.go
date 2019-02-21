package detailed

import (
	"github.com/weaveworks/scope/report"
)

func isCommand(key string) bool {
	return key == report.Cmdline || key == report.DockerContainerCommand
}

func censorNodeSummary(s *NodeSummary, cfg report.CensorConfig) {
	if cfg.HideEnvironmentVariables {
		tables := []report.Table{}
		for _, t := range s.Tables {
			if t.ID != report.DockerEnvPrefix {
				tables = append(tables, t)
			}
		}
		s.Tables = tables
	}
	if cfg.HideCommandLineArguments {
		for r := range s.Metadata {
			if isCommand(s.Metadata[r].ID) {
				s.Metadata[r].Value = report.StripCommandArgs(s.Metadata[r].Value)
			}
		}
	}
}

// CensorNode ...
func CensorNode(n Node, cfg report.CensorConfig) Node {
	censorNodeSummary(&n.NodeSummary, cfg)
	return n
}

// CensorNodeSummaries ...
func CensorNodeSummaries(ns NodeSummaries, cfg report.CensorConfig) NodeSummaries {
	for key := range ns {
		n := ns[key]
		censorNodeSummary(&n, cfg)
		ns[key] = n
	}
	return ns
}
