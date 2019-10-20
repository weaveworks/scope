package report

// Backwards-compatibility: code to read older reports and convert

import (
	"sort"
	"strings"
	"time"

	"github.com/ugorji/go/codec"
)

// For backwards-compatibility with probes that sent a map of latestControls data
type bcNode struct {
	Node
	LatestControls map[string]nodeControlDataLatestEntry `json:"latestControls,omitempty"`
	OldStringMap   map[string]oldStringEntry             `json:"latest,omitempty"`
	Counters       map[string]int                        `json:"counters,omitempty"`
}

type nodeControlDataLatestEntry struct {
	Timestamp time.Time       `json:"timestamp"`
	Value     nodeControlData `json:"value"`
}

type nodeControlData struct {
	Dead bool `json:"dead"`
}

type oldStringEntry struct {
	// Timestamp time.Time `json:"timestamp"`  // we don't look at the individual timestamps
	Value string `json:"value"`
}

// CodecDecodeSelf implements codec.Selfer
func (n *Node) CodecDecodeSelf(decoder *codec.Decoder) {
	var in bcNode
	decoder.Decode(&in)
	*n = in.Node
	if len(in.OldStringMap) > 0 {
		n.Latest = make(StringLatestMap, 0, len(in.OldStringMap))
		for key, data := range in.OldStringMap {
			n.Latest = append(n.Latest, stringLatestEntry{key: key, value: data.Value})
		}
		sort.Sort(n.Latest)
	}
	if len(in.LatestControls) > 0 {
		// Convert the map into a delimited string
		cs := make([]string, 0, len(in.LatestControls))
		for name, v := range in.LatestControls {
			if !v.Value.Dead {
				cs = append(cs, name)
			}
		}
		n.Latest = n.Latest.Set(NodeActiveControls, strings.Join(cs, ScopeDelim))
	}
	// Counters were not generated in the Scope probe, but decode them here in case a plugin used them.
	for k, v := range in.Counters {
		*n = n.WithCounter(k, v)
	}
}

type _Node Node // just so we don't recurse inside CodecEncodeSelf

// CodecEncodeSelf implements codec.Selfer
func (n *Node) CodecEncodeSelf(encoder *codec.Encoder) {
	encoder.Encode((*_Node)(n))
}

// Upgrade returns a new report based on a report received from the old probe.
//
func (r Report) Upgrade() Report {
	return r.upgradePodNodes().upgradeNamespaces().upgradeDNSRecords()
}

func (r Report) upgradePodNodes() Report {
	// At the same time the probe stopped reporting replicasets,
	// it also started reporting deployments as pods' parents
	if len(r.ReplicaSet.Nodes) == 0 {
		return r
	}

	// For each pod, we check for any replica sets, and merge any deployments they point to
	// into a replacement Parents value.
	nodes := Nodes{}
	for podID, pod := range r.Pod.Nodes {
		if replicaSetIDs, ok := pod.Parents.Lookup(ReplicaSet); ok {
			newParents := pod.Parents.Delete(ReplicaSet)
			for _, replicaSetID := range replicaSetIDs {
				if replicaSet, ok := r.ReplicaSet.Nodes[replicaSetID]; ok {
					if deploymentIDs, ok := replicaSet.Parents.Lookup(Deployment); ok {
						newParents = newParents.Add(Deployment, deploymentIDs)
					}
				}
			}
			// newParents contains a copy of the current parents without replicasets,
			// PruneParents().WithParents() ensures replicasets are actually deleted
			pod = pod.PruneParents().WithParents(newParents)
		}
		nodes[podID] = pod
	}
	r.Pod.Nodes = nodes

	return r
}

func (r Report) upgradeNamespaces() Report {
	if len(r.Namespace.Nodes) > 0 {
		return r
	}

	namespaces := map[string]struct{}{}
	for _, t := range []Topology{r.Pod, r.Service, r.Deployment, r.DaemonSet, r.StatefulSet, r.CronJob} {
		for _, n := range t.Nodes {
			if state, ok := n.Latest.Lookup(KubernetesState); ok && state == "deleted" {
				continue
			}
			if namespace, ok := n.Latest.Lookup(KubernetesNamespace); ok {
				namespaces[namespace] = struct{}{}
			}
		}
	}

	nodes := make(Nodes, len(namespaces))
	for ns := range namespaces {
		// Namespace ID:
		// Probes did not use to report namespace ids, but since creating a report node requires an id,
		// the namespace name, which is unique, is passed to `MakeNamespaceNodeID`
		namespaceID := MakeNamespaceNodeID(ns)
		nodes[namespaceID] = MakeNodeWith(namespaceID, map[string]string{KubernetesName: ns})
	}
	r.Namespace.Nodes = nodes

	return r
}

func (r Report) upgradeDNSRecords() Report {
	if len(r.DNS) > 0 {
		return r
	}
	dns := make(DNSRecords)
	for endpointID, endpoint := range r.Endpoint.Nodes {
		_, addr, _, ok := ParseEndpointNodeID(endpointID)
		snoopedNames, foundS := endpoint.Sets.Lookup(SnoopedDNSNames)
		reverseNames, foundR := endpoint.Sets.Lookup(ReverseDNSNames)
		if ok && (foundS || foundR) {
			// Add address and names to report-level map
			if existing, found := dns[addr]; found {
				var sUnchanged, rUnchanged bool
				snoopedNames, sUnchanged = snoopedNames.Merge(existing.Forward)
				reverseNames, rUnchanged = reverseNames.Merge(existing.Reverse)
				if sUnchanged && rUnchanged {
					continue
				}
			}
			dns[addr] = DNSRecord{Forward: snoopedNames, Reverse: reverseNames}
		}
	}
	r.DNS = dns
	return r
}
