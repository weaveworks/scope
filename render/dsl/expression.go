package dsl

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strings"

	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// Evaluator describes a monadic transformer of a topology. It is expected
// that the topology contains heterogeneous nodes.
type Evaluator interface {
	Eval(report.Topology) report.Topology
}

// Expression is a single evaluator.
type Expression struct {
	selector
	transformer
}

// Eval implements Evaluator.
func (e Expression) Eval(tpy report.Topology) report.Topology {
	return e.transformer(tpy, e.selector(tpy))
}

// Expressions is an ordered collection of expressions.
type Expressions []Expression

// Eval implements Evaluator.
func (e Expressions) Eval(tpy report.Topology) report.Topology {
	for _, expr := range e {
		tpy = expr.Eval(tpy)
	}
	return tpy
}

type selector func(report.Topology) []string

type transformer func(report.Topology, []string) report.Topology

func selectAll(tpy report.Topology) []string {
	out := make([]string, 0, len(tpy.NodeMetadatas))
	for id := range tpy.NodeMetadatas {
		out = append(out, id)
	}
	log.Printf("select ALL: %d", len(out))
	return out
}

func selectConnected(tpy report.Topology) []string {
	degree := map[string]int{}
	for src, dsts := range tpy.Adjacency {
		a, ok := report.ParseAdjacencyID(src)
		if !ok {
			panic(src)
		}
		degree[a] += len(dsts)
		for _, dst := range dsts {
			degree[dst]++
		}
	}
	out := []string{}
	for id := range tpy.NodeMetadatas {
		if degree[id] > 0 {
			out = append(out, id)
		}
	}
	log.Printf("select CONNECTED: %d", len(out))
	return out
}

func selectNonlocal(tpy report.Topology) []string {
	local := report.Networks{}
	for _, md := range tpy.NodeMetadatas {
		for k, v := range md.Metadata {
			if k == host.LocalNetworks {
				local = append(local, render.ParseNetworks(v)...)
			}
		}
	}
	out := []string{}
	for id, md := range tpy.NodeMetadatas {
		if addr, ok := md.Metadata[endpoint.Addr]; ok {
			if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
				out = append(out, id) // valid addr metadata key, nonlocal
				continue
			}
		}
		if _, addr, ok := report.ParseAddressNodeID(id); ok {
			if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
				out = append(out, id) // valid address node ID, nonlocal
				continue
			}
		}
		if _, addr, _, ok := report.ParseEndpointNodeID(id); ok {
			if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
				out = append(out, id) // valid endpoint node ID, nonlocal
				continue
			}
		}
	}
	log.Printf("select NONLOCAL: %d", len(out))
	return out
}

func selectLike(s string) selector {
	re, err := regexp.Compile(s)
	if err != nil {
		log.Printf("select LIKE %q: %v", s, err)
		re = regexp.MustCompile("")
	}
	return func(tpy report.Topology) []string {
		out := []string{}
		for id := range tpy.NodeMetadatas {
			if re.MatchString(id) {
				out = append(out, id)
			}
		}
		log.Printf("select LIKE %q: %d", s, len(out))
		return out
	}
}

func selectWith(s string) selector {
	var k, v string
	if fields := strings.SplitN(s, "=", 2); len(fields) == 1 {
		k = strings.TrimSpace(fields[0])
	} else if len(fields) == 2 {
		k, v = strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
	}

	return func(tpy report.Topology) []string {
		out := []string{}
		for id, md := range tpy.NodeMetadatas {
			if vv, ok := md.Metadata[k]; ok {
				if v == "" || (v != "" && v == vv) {
					out = append(out, id)
				}
			}
		}
		log.Printf("select WITH %q: %d", s, len(out))
		return out
	}
}

func selectNot(s selector) selector {
	return func(tpy report.Topology) []string {
		set := map[string]struct{}{}
		for _, id := range s(tpy) {
			set[id] = struct{}{}
		}
		out := []string{}
		for id := range tpy.NodeMetadatas {
			if _, ok := set[id]; ok {
				continue // selected by that one -> not by this one
			}
			out = append(out, id)
		}
		log.Printf("select NOT: %d", len(out))
		return out
	}
}

const highlightKey = "_highlight"

func transformHighlight(tpy report.Topology, ids []string) report.Topology {
	for _, id := range ids {
		tpy.NodeMetadatas[id] = tpy.NodeMetadatas[id].Merge(report.MakeNodeMetadataWith(map[string]string{highlightKey: "true"}))
	}
	log.Printf("transform HIGHLIGHT %d: OK", len(ids))
	return tpy
}

func transformRemove(tpy report.Topology, ids []string) report.Topology {
	toRemove := map[string]struct{}{}
	for _, id := range ids {
		toRemove[id] = struct{}{}
	}
	out := report.MakeTopology()
	for id := range tpy.NodeMetadatas {
		if _, ok := toRemove[id]; ok {
			continue
		}
		cp(out, tpy, id)
	}
	clean(out, toRemove)
	log.Printf("transform REMOVE %d: in %d, out %d", len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
	return out
}

func transformShowOnly(tpy report.Topology, ids []string) report.Topology {
	out := report.MakeTopology()
	for _, id := range ids {
		if _, ok := tpy.NodeMetadatas[id]; !ok {
			continue
		}
		cp(out, tpy, id)
	}
	log.Printf("transform SHOWONLY %d: in %d, out %d", len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
	return out
}

func transformMerge(tpy report.Topology, ids []string) report.Topology {
	name := fmt.Sprintf("%x", rand.Int31())
	mapped := map[string]string{}
	for _, id := range ids {
		mapped[id] = name
	}
	out := report.MakeTopology()
	for id := range tpy.NodeMetadatas {
		if dstID, ok := mapped[id]; ok {
			merge(out, dstID, tpy, id, mapped)
		} else {
			cp(out, tpy, id)
		}
	}
	log.Printf("transform MERGE %d: in %d, out %d", len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
	return out
}

func transformGroupBy(s string) transformer {
	keys := []string{}
	for _, key := range strings.Split(s, ",") {
		keys = append(keys, strings.TrimSpace(key))
	}

	return func(tpy report.Topology, ids []string) report.Topology {
		set := map[string]struct{}{}
		for _, id := range ids {
			set[id] = struct{}{}
		}

		// Identify all nodes that should be grouped.
		mapped := map[string]string{} // src ID: dst ID
		for id, md := range tpy.NodeMetadatas {
			if _, ok := set[id]; !ok {
				continue // not selected
			}

			parts := []string{}
			for _, key := range keys {
				if val, ok := md.Metadata[key]; ok {
					parts = append(parts, fmt.Sprintf("%s-%s", key, val))
				}
			}
			if len(parts) < len(keys) {
				continue // didn't match all required keys
			}

			dstID := strings.Join(parts, "-")
			mapped[id] = dstID
		}

		// Walk nodes again, merging those that should be grouped.
		out := report.MakeTopology()
		for id := range tpy.NodeMetadatas {
			if dstID, ok := mapped[id]; ok {
				merge(out, dstID, tpy, id, mapped)
			} else {
				cp(out, tpy, id)
			}
		}

		log.Printf("transform GROUPBY %v %d: in %d, out %d", keys, len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
		return out
	}
}

func transformJoin(key string) transformer {
	return func(tpy report.Topology, ids []string) report.Topology {
		// key is e.g. host_node_id.
		// Collect the set of represented values.
		values := map[string]report.NodeMetadata{}
		for _, md := range tpy.NodeMetadatas {
			for k, v := range md.Metadata {
				if k == key {
					values[v] = report.MakeNodeMetadata() // gather later
				}
			}
		}

		// Next, gather the metadata from nodes in the set.
		for id, md := range tpy.NodeMetadatas {
			if found, ok := values[id]; ok {
				values[id] = found.Merge(md) // gather
			}
		}

		// Finally, join that metadata to referential nodes.
		// And delete the referenced nodes.
		out := report.MakeTopology()
		for id, md := range tpy.NodeMetadatas {
			if _, ok := values[id]; ok {
				continue // delete
			}
			cp(out, tpy, id) // copy node
			for k, v := range md.Metadata {
				if k == key {
					md = md.Merge(values[v]) // join metadata
				}
			}
			out.NodeMetadatas[id] = md // write
		}

		log.Printf("transform JOIN %v %d: in %d, out %d", key, len(ids), len(tpy.NodeMetadatas), len(out.NodeMetadatas))
		return out
	}
}

func cp(dst report.Topology, src report.Topology, id string) {
	adjacencyID := report.MakeAdjacencyID(id)
	dst.Adjacency[adjacencyID] = src.Adjacency[adjacencyID]

	for _, otherID := range dst.Adjacency[id] {
		edgeID := report.MakeEdgeID(id, otherID)
		dst.EdgeMetadatas[edgeID] = src.EdgeMetadatas[edgeID]
	}

	dst.NodeMetadatas[id] = src.NodeMetadatas[id]
}

func clean(dst report.Topology, toRemove map[string]struct{}) {
	for srcAdjacencyID, dstIDs := range dst.Adjacency {
		newIDs := report.IDList{}
		for _, dstID := range dstIDs {
			if _, ok := toRemove[dstID]; ok {
				continue // can't be a dst anymore
			}
			newIDs = append(newIDs, dstID)
		}
		if len(newIDs) <= 0 {
			delete(dst.Adjacency, srcAdjacencyID) // all dsts are gone, so rm src
			continue
		}
		dst.Adjacency[srcAdjacencyID] = newIDs // overwrite
	}

	for id := range toRemove {
		delete(dst.NodeMetadatas, id)
	}

	for edgeID := range dst.EdgeMetadatas {
		srcNodeID, dstNodeID, ok := report.ParseEdgeID(edgeID)
		if !ok {
			continue // panic
		}
		if _, ok := toRemove[srcNodeID]; ok {
			delete(dst.EdgeMetadatas, edgeID)
		}
		if _, ok := toRemove[dstNodeID]; ok {
			delete(dst.EdgeMetadatas, edgeID)
		}
	}
}

func merge(dst report.Topology, dstID string, src report.Topology, srcID string, mapped map[string]string) {
	// We gonna take srcID from src and merge it into dstID in dst. That's
	// like renaming the node, so we gotta update the adjacency lists. Both
	// outgoing *and* incoming links!
	dstAdjacencyID := report.MakeAdjacencyID(dstID)
	srcAdjacencyID := report.MakeAdjacencyID(srcID)

	// Merge the src's adjacency list into the dst topology.
	dst.Adjacency[dstAdjacencyID] = dst.Adjacency[dstAdjacencyID].Merge(src.Adjacency[srcAdjacencyID])

	// Update any dst adjacencies from the old ID to the new ID.
	for existingSrcAdjacencyID, existingDstIDs := range dst.Adjacency {
		for i, existingDstID := range existingDstIDs {
			if newDstID, ok := mapped[existingDstID]; ok {
				existingDstIDs[i] = newDstID
			}
		}
		dst.Adjacency[existingSrcAdjacencyID] = existingDstIDs
	}

	// Update the EdgeMetadatas to have the new IDs.
	for _, otherID := range src.Adjacency[srcAdjacencyID] {
		oldEdgeID := report.MakeEdgeID(srcID, otherID)
		newEdgeID := report.MakeEdgeID(dstID, otherID)
		dst.EdgeMetadatas[newEdgeID] = dst.EdgeMetadatas[newEdgeID].Merge(src.EdgeMetadatas[oldEdgeID])
	}

	// Merge the src node metadata into the dst node metadata.
	md, ok := dst.NodeMetadatas[dstID]
	if !ok {
		md = report.MakeNodeMetadata()
	}
	md = md.Merge(src.NodeMetadatas[srcID])
	dst.NodeMetadatas[dstID] = md
}
