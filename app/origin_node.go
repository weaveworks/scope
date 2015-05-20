package main

import "github.com/weaveworks/scope/report"

// OriginNode is a node in the originating report topology. It's a process ID
// or network host. It's used by the /api/topology/{topology}/{nodeID} handler
// to generate detailed information. One node from a rendered topology may
// have multiple origin nodes.
type OriginNode struct {
	Table report.Table
}

func getOriginNode(t report.Topology, id string) (OriginNode, bool) {
	node, ok := t.NodeMetadatas[id]
	if !ok {
		return OriginNode{}, false
	}

	// The node represents different actual things depending on the topology.
	// So we deduce what it is, based on the metadata.
	if _, ok := node["pid"]; ok {
		return originNodeForProcess(node), true
	}

	// Assume network host. Could strengthen this guess by adding a
	// special key in the probe spying procedure.
	return originNodeForNetworkHost(node), true
}

func originNodeForProcess(node report.NodeMetadata) OriginNode {
	rows := []report.Row{
		{Key: "Host", ValueMajor: node["domain"], ValueMinor: ""},
		{Key: "PID", ValueMajor: node["pid"], ValueMinor: ""},
		{Key: "Process name", ValueMajor: node["name"], ValueMinor: ""},
	}
	if val, ok := node["docker_id"]; ok {
		rows = append(rows, report.Row{Key: "Container ID", ValueMajor: val, ValueMinor: ""})
	}
	if val, ok := node["docker_name"]; ok {
		rows = append(rows, report.Row{Key: "Container name", ValueMajor: val, ValueMinor: ""})
	}
	if val, ok := node["cgroup"]; ok {
		rows = append(rows, report.Row{Key: "cgroup", ValueMajor: val, ValueMinor: ""})
	}
	return OriginNode{
		Table: report.Table{
			Title:   "Origin Process",
			Numeric: false,
			Rows:    rows,
		},
	}
}

func originNodeForNetworkHost(node report.NodeMetadata) OriginNode {
	rows := []report.Row{
		{"Hostname", node["name"], ""},
	}
	return OriginNode{
		Table: report.Table{
			Title:   "Origin Host",
			Numeric: false,
			Rows:    rows,
		},
	}
}
