package dsl

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
)

// Evaluator describes a monadic transformer of a RenderableNodes.
type Evaluator interface {
	Eval(render.RenderableNodes) render.RenderableNodes
}

// Expression is a single evaluator.
type Expression struct {
	selector
	transformer
}

// Eval implements Evaluator.
func (e Expression) Eval(rns render.RenderableNodes) render.RenderableNodes {
	return e.transformer(rns, e.selector(rns))
}

// Expressions is an ordered collection of expressions.
type Expressions []Expression

// Eval implements Evaluator.
func (e Expressions) Eval(rns render.RenderableNodes) render.RenderableNodes {
	for _, expr := range e {
		rns = expr.Eval(rns)
	}
	return rns
}

type selector func(render.RenderableNodes) []string

type transformer func(render.RenderableNodes, []string) render.RenderableNodes

func selectAll(rns render.RenderableNodes) []string {
	out := make([]string, 0, len(rns))
	for id := range rns {
		out = append(out, id)
	}
	//log.Printf("select ALL: %d", len(out))
	return out
}

func selectConnected(rns render.RenderableNodes) []string {
	degree := map[string]int{}
	for src, rn := range rns {
		degree[src] += len(rn.Adjacency)
		for _, dst := range rn.Adjacency {
			degree[dst]++
		}
	}
	out := []string{}
	for id := range rns {
		if degree[id] > 0 {
			out = append(out, id)
		}
	}
	return out
}

func selectNonlocal(rns render.RenderableNodes) []string {
	local := report.Networks{}
	for _, rn := range rns {
		for k, v := range rn.Metadata {
			if k == host.LocalNetworks {
				local = append(local, report.ParseNetworks(v)...)
			}
		}
	}
	out := []string{}
	for id, rn := range rns {
		if addr, ok := rn.Metadata[endpoint.Addr]; ok {
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
	//log.Printf("select NONLOCAL: %d", len(out))
	return out
}

func selectLike(regex string) selector {
	re, err := regexp.Compile(regex)
	if err != nil {
		//log.Printf("select LIKE %q: %v", s, err)
		re = regexp.MustCompile("")
	}
	return func(rns render.RenderableNodes) []string {
		out := []string{}
		for id := range rns {
			if re.MatchString(id) {
				out = append(out, id)
			}
		}
		//log.Printf("select LIKE %q: %d", s, len(out))
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
	return func(rns render.RenderableNodes) []string {
		out := []string{}
		for id, md := range rns {
			if vv, ok := md.Metadata[k]; ok {
				if v == "" || (v != "" && v == vv) {
					out = append(out, id)
				}
			}
		}
		//log.Printf("select WITH %q: %d", s, len(out))
		return out
	}
}

func selectNot(s selector) selector {
	return func(rns render.RenderableNodes) []string {
		set := map[string]struct{}{}
		for _, id := range s(rns) {
			set[id] = struct{}{}
		}
		out := []string{}
		for id := range rns {
			if _, ok := set[id]; ok {
				continue // selected by that one -> not by this one
			}
			out = append(out, id)
		}
		//log.Printf("select NOT: %d", len(out))
		return out
	}
}

const highlightKey = "_highlight"

func transformHighlight(rns render.RenderableNodes, ids []string) render.RenderableNodes {
	for _, id := range ids {
		rn := rns[id]
		rn.Node.Metadata[highlightKey] = "true"
		rns[id] = rn
	}
	//log.Printf("transform HIGHLIGHT %d: OK", len(ids))
	return rns
}

func transformRemove(rns render.RenderableNodes, ids []string) render.RenderableNodes {
	toRemove := map[string]struct{}{}
	for _, id := range ids {
		toRemove[id] = struct{}{}
	}
	out := render.RenderableNodes{}
	for id := range rns {
		if _, ok := toRemove[id]; ok {
			continue
		}
		cp(out, rns, id)
	}
	clean(out, toRemove)
	//log.Printf("transform REMOVE %d: in %d, out %d", len(ids), len(rns), len(out.NodeMetadatas))
	return out
}

func transformShowOnly(rns render.RenderableNodes, ids []string) render.RenderableNodes {
	out := render.RenderableNodes{}
	for _, id := range ids {
		cp(out, rns, id)
	}
	toRemove := map[string]struct{}{}
	for id := range rns {
		if _, ok := out[id]; !ok {
			toRemove[id] = struct{}{}
		}
	}
	clean(out, toRemove)
	//log.Printf("transform SHOWONLY %d: in %d, out %d", len(ids), len(rns), len(out.NodeMetadatas))
	return out
}

func transformMerge(newname string) transformer {
	return func(rns render.RenderableNodes, ids []string) render.RenderableNodes {
		mapped := map[string]string{}
		toRemove := map[string]struct{}{}
		for _, id := range ids {
			mapped[id] = newname
			toRemove[id] = struct{}{}
		}
		out := render.RenderableNodes{}
		for id := range rns {
			if dstID, ok := mapped[id]; ok {
				merge(out, dstID, rns, id)
			} else {
				cp(out, rns, id)
			}
		}
		shift(out, mapped)
		clean(out, toRemove)
		//log.Printf("transform MERGE %d: in %d, out %d", len(ids), len(rns), len(out.NodeMetadatas))
		return out
	}
}

// transformGroupBy takes a key, and merges all nodes who share the same value
// for that key. It ignores nodes that don't have that key.
func transformGroupBy(s string) transformer {
	keys := []string{}
	for _, key := range strings.Split(s, ",") {
		keys = append(keys, strings.TrimSpace(key))
	}

	return func(rns render.RenderableNodes, ids []string) render.RenderableNodes {
		set := map[string]struct{}{}
		for _, id := range ids {
			set[id] = struct{}{}
		}

		// Identify all nodes that should be grouped.
		mapped := map[string]string{} // src ID: dst ID
		toRemove := map[string]struct{}{}
		for id, md := range rns {
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
			toRemove[id] = struct{}{}
		}

		// Walk nodes again, merging those that should be grouped.
		out := render.RenderableNodes{}
		for id := range rns {
			if dstID, ok := mapped[id]; ok {
				merge(out, dstID, rns, id)
			} else {
				cp(out, rns, id)
			}
		}
		shift(out, mapped)
		clean(out, toRemove)

		//log.Printf("transform GROUPBY %v %d: in %d, out %d", keys, len(ids), len(rns), len(out.NodeMetadatas))
		return out
	}
}

// transformJoin takes a key whose value is (expected to be) a node ID.
// Basically, a foreign key. For every unique value (i.e. node) it finds, it
// copies the foreign node's metadata into every node that had the
// corresponding foreign key, and then deletes the foreign node.
//
// It's kind of like flattening the foreign nodes into all the nodes that
// point to them, one-to-many.
func transformJoin(key string) transformer {
	return func(rns render.RenderableNodes, ids []string) render.RenderableNodes {
		// key is e.g. host_node_id, value is a valid node ID.
		// Collect the set of represented values (node IDs).
		values := map[string]report.Node{}
		toRemove := map[string]struct{}{}
		for _, rn := range rns {
			for k, v := range rn.Node.Metadata {
				if k == key {
					values[v] = report.MakeNode() // gather later
					toRemove[v] = struct{}{}
				}
			}
		}

		// Next, gather the metadata from nodes in the set.
		for id, rn := range rns {
			if found, ok := values[id]; ok {
				values[id] = found.Merge(rn.Node) // gather
			}
		}

		// Finally, join that metadata to referential nodes.
		// And delete the referenced nodes.
		out := render.RenderableNodes{}
		for id, rn := range rns {
			if _, ok := values[id]; ok {
				continue // delete the foreign nodes
			}
			cp(out, rns, id) // copy node
			for k, v := range rn.Node.Metadata {
				if k == key {
					rn.Node = rn.Node.Merge(values[v]) // join metadata
				}
			}
			out[id] = rn // write
		}
		clean(out, toRemove)

		//log.Printf("transform JOIN %v %d: in %d, out %d", key, len(ids), len(rns), len(out.NodeMetadatas))
		return out
	}
}

func cp(dst render.RenderableNodes, src render.RenderableNodes, id string) {
	dst[id] = src[id].Copy()

	// Every transform that calls cp must call clean at the end, to remove
	// dangling (uncopied) nodes from adjacency lists and edge metadatas.
}

func merge(dst render.RenderableNodes, dstID string, src render.RenderableNodes, srcID string) {
	dstNode, ok := dst[dstID]
	if !ok {
		dstNode = render.NewRenderableNode(dstID)
	}
	dst[dstID] = dstNode.Merge(src[srcID])

	// Every transform that calls merge must call shift at the end, to update
	// adjacency lists and edge metadatas.
}

func clean(dst render.RenderableNodes, toRemove map[string]struct{}) {
	for id, rn := range dst {
		// Clean out all the orphans from the adjacency list.
		newAdjacency := report.IDList{}
		for _, otherID := range rn.Node.Adjacency {
			if _, ok := toRemove[otherID]; ok {
				continue // can't be a dst anymore
			}
			newAdjacency = newAdjacency.Add(otherID)
		}
		rn.Node.Adjacency = newAdjacency

		// Clean out all the orphans from the edges.
		newEdges := report.EdgeMetadatas{}
		for otherID, edge := range rn.Node.Edges {
			if _, ok := toRemove[otherID]; ok {
				continue // can't be an edge anymore
			}
			newEdges[otherID] = edge
		}
		rn.Node.Edges = newEdges

		dst[id] = rn
	}

	// Just to be safe.
	for id := range toRemove {
		delete(dst, id)
	}
}

func shift(dst render.RenderableNodes, mapping map[string]string) {
	// We've got a mapping of old IDs to new IDs. Any adjacency targeting an
	// old ID should be updated to the new ID.
	for id, rn := range dst {
		newAdjacency := report.IDList{}
		for _, otherID := range rn.Node.Adjacency {
			if mappedID, ok := mapping[otherID]; ok {
				otherID = mappedID // just shift it on over
			}
			newAdjacency = newAdjacency.Add(otherID) // this will dedupe
		}
		rn.Node.Adjacency = newAdjacency

		newEdges := report.EdgeMetadatas{}
		for otherID, edge := range rn.Node.Edges {
			if mappedID, ok := mapping[otherID]; ok {
				otherID = mappedID
			}
			newEdges[otherID] = newEdges[otherID].Merge(edge) // important to merge here
		}
		rn.Node.Edges = newEdges

		dst[id] = rn
	}
}
