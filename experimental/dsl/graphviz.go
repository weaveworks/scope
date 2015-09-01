package main

import (
	"fmt"
	"io"
	"sort"

	"github.com/weaveworks/scope/probe/host"

	"github.com/weaveworks/scope/report"
)

func dot(w io.Writer, tpy report.Topology) {
	var nodes []string
	for id, md := range tpy.NodeMetadatas {
		label := id + "\n"

		var lines []string
		var highlight bool
		for k, v := range md.Metadata {
			if k == host.LocalNetworks {
				continue
			}
			if k == highlightKey {
				highlight = true
				continue
			}
			lines = append(lines, fmt.Sprintf("%s=%s", k, v))
		}
		sort.Strings(lines)
		for _, line := range lines {
			label += "\n" + line
		}

		fillcolor := "white"
		if highlight {
			fillcolor = "yellow"
		}

		nodes = append(nodes, fmt.Sprintf("%q [label=%q fillcolor=%s];", id, label, fillcolor))
	}

	var edges []string
	for src, dsts := range tpy.Adjacency {
		a, ok := report.ParseAdjacencyID(src)
		if !ok {
			panic(src)
		}
		for _, dst := range dsts {
			edges = append(edges, fmt.Sprintf("%q -> %q;", a, dst))
		}
	}

	sort.Strings(nodes)
	sort.Strings(edges)

	fmt.Fprintf(w, "digraph G {\n")
	fmt.Fprintf(w, "\tgraph [ overlap=false, mode=hier ];\n")
	fmt.Fprintf(w, "\tnode [ shape=rect, style=filled ];\n")
	fmt.Fprintf(w, "\toutputorder=edgesfirst;\n")
	fmt.Fprintf(w, "\n")
	for _, line := range nodes {
		fmt.Fprintf(w, "\t%s\n", line)
	}
	fmt.Fprintf(w, "\n")
	for _, line := range edges {
		fmt.Fprintf(w, "\t%s\n", line)
	}
	fmt.Fprintf(w, "}\n")
}
