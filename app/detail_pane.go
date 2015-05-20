package main

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/weaveworks/scope/report"
)

func makeDetailed(
	n report.RenderableNode,
	originHostLookup func(string) (OriginHost, bool),
	originNodeLookup func(string) (OriginNode, bool),
) report.DetailedNode {
	tables := []report.Table{{
		Title:   "Connections",
		Numeric: true,
		Rows: []report.Row{
			// TODO omit these rows if there's no data?
			{"TCP connections", strconv.FormatInt(int64(n.Metadata[report.KeyMaxConnCountTCP]), 10), ""},
			{"Bytes ingress", strconv.FormatInt(int64(n.Metadata[report.KeyBytesIngress]), 10), ""},
			{"Bytes egress", strconv.FormatInt(int64(n.Metadata[report.KeyBytesEgress]), 10), ""},
		},
	}}

	// Note that a RenderableNode may be the result of merge operation(s), and
	// so may have multiple origin hosts and nodes.

outer:
	for _, id := range n.OriginNodes {
		// Origin node IDs in e.g. the process topology are actually network
		// n-tuples. (The process topology is actually more like a network
		// n-tuple topology.) So we can have multiple IDs mapping to the same
		// process. There are several ways to dedupe that, but here we take
		// the lazy way and do simple equivalence of the resulting table.
		node, ok := originNodeLookup(id)
		if !ok {
			node = unknownOriginNode(id)
		}
		for _, table := range tables {
			if reflect.DeepEqual(table, node.Table) {
				continue outer
			}
		}
		tables = append(tables, node.Table)
	}

	for _, id := range n.OriginHosts {
		host, ok := originHostLookup(id)
		if !ok {
			host = unknownOriginHost(id)
		}
		tables = append(tables, report.Table{
			Title:   "Origin Host",
			Numeric: false,
			Rows: []report.Row{
				{"Hostname", host.Hostname, ""},
				{"Load", fmt.Sprintf("%.2f %.2f %.2f", host.LoadOne, host.LoadFive, host.LoadFifteen), ""},
				{"OS", host.OS, ""},
				//{"Addresses", strings.Join(host.Addresses, ", "), ""},
				{"ID", id, ""},
			},
		})
	}

	return report.DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Tables:     tables,
	}
}

func unknownOriginHost(id string) OriginHost {
	return OriginHost{
		Hostname:    fmt.Sprintf("[%s]", id),
		OS:          "unknown",
		Addresses:   []string{},
		LoadOne:     0.0,
		LoadFive:    0.0,
		LoadFifteen: 0.0,
	}
}

func unknownOriginNode(id string) OriginNode {
	return OriginNode{
		Table: report.Table{
			Title:   "Origin Node",
			Numeric: false,
			Rows: []report.Row{
				{"ID", id, ""},
			},
		},
	}
}
