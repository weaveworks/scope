package detailed

import (
	"github.com/weaveworks/scope/report"
)

// CensorNode ...
func CensorNode(n Node, cfg report.CensorConfig) Node {
	return n
}

// CensorNodeSummaries ...
func CensorNodeSummaries(n NodeSummaries, cfg report.CensorConfig) NodeSummaries {
	return n
}
