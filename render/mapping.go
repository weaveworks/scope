package render

import (
	"fmt"
	"net"
	"regexp"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	UncontainedID    = "uncontained"
	UncontainedMajor = "Uncontained"

	TheInternetID      = "theinternet"
	IncomingInternetID = "in-" + TheInternetID
	OutgoingInternetID = "out-" + TheInternetID
	InboundMajor       = "The Internet"
	OutboundMajor      = "The Internet"
	InboundMinor       = "Inbound connections"
	OutboundMinor      = "Outbound connections"

	ipsKey = "ips"

	// Topology for pseudo-nodes and IPs so we can differentiate them at the end
	Pseudo = "pseudo"
	IP     = "IP"
)

// MapFunc is anything which can take an arbitrary Node and
// return a set of other Nodes.
//
// If the output is empty, the node shall be omitted from the rendered topology.
type MapFunc func(report.Node, report.Networks) report.Nodes

// NewDerivedNode makes a node based on node, but with a new ID
func NewDerivedNode(id string, node report.Node) report.Node {
	return node.WithID(id).WithChildren(report.MakeNodeSet(node)).PruneParents()
}

// NewDerivedPseudoNode makes a new pseudo node with the node as a child
func NewDerivedPseudoNode(id string, node report.Node) report.Node {
	return node.WithID(id).WithTopology(Pseudo).WithChildren(report.MakeNodeSet(node)).PruneParents()
}

func theInternetNode(m report.Node) report.Node {
	// emit one internet node for incoming, one for outgoing
	if len(m.Adjacency) > 0 {
		return NewDerivedPseudoNode(IncomingInternetID, m)
	}
	return NewDerivedPseudoNode(OutgoingInternetID, m)
}

// MapEndpoint2IP maps endpoint nodes to their IP address, for joining
// with container nodes.  We drop endpoint nodes with pids, as they
// will be joined to containers through the process topology, and we
// don't want to double count edges.
func MapEndpoint2IP(m report.Node, local report.Networks) report.Nodes {
	// Don't include procspied connections, to prevent double counting
	_, ok := m.Latest.Lookup(endpoint.Procspied)
	if ok {
		return report.Nodes{}
	}
	scope, addr, port, ok := report.ParseEndpointNodeID(m.ID)
	if !ok {
		return report.Nodes{}
	}
	if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
		return report.Nodes{TheInternetID: theInternetNode(m)}
	}

	// We don't always know what port a container is listening on, and
	// container-to-container communications can be unambiguously identified
	// without ports. OTOH, connections to the host IPs which have been port
	// mapped to a container can only be unambiguously identified with the port.
	// So we need to emit two nodes, for two different cases.
	id := report.MakeScopedEndpointNodeID(scope, addr, "")
	idWithPort := report.MakeScopedEndpointNodeID(scope, addr, port)
	return report.Nodes{
		id:         NewDerivedNode(id, m).WithTopology(IP),
		idWithPort: NewDerivedNode(idWithPort, m).WithTopology(IP),
	}
}

var portMappingMatch = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}):([0-9]+)->([0-9]+)/tcp`)

// MapContainer2IP maps container nodes to their IP addresses (outputs
// multiple nodes).  This allows container to be joined directly with
// the endpoint topology.
func MapContainer2IP(m report.Node, _ report.Networks) report.Nodes {
	containerID, ok := m.Latest.Lookup(docker.ContainerID)
	if !ok {
		return report.Nodes{}
	}

	result := report.Nodes{}
	if addrs, ok := m.Sets.Lookup(docker.ContainerIPsWithScopes); ok {
		for _, addr := range addrs {
			scope, addr, ok := report.ParseAddressNodeID(addr)
			if !ok {
				continue
			}
			id := report.MakeScopedEndpointNodeID(scope, addr, "")
			result[id] = NewDerivedNode(id, m).
				WithTopology(IP).
				WithLatests(map[string]string{docker.ContainerID: containerID}).
				WithCounters(map[string]int{ipsKey: 1})

		}
	}

	// Also output all the host:port port mappings (see above comment).
	// In this case we assume this doesn't need a scope, as they are for host IPs.
	ports, _ := m.Sets.Lookup(docker.ContainerPorts)
	for _, portMapping := range ports {
		if mapping := portMappingMatch.FindStringSubmatch(portMapping); mapping != nil {
			ip, port := mapping[1], mapping[2]
			id := report.MakeScopedEndpointNodeID("", ip, port)
			result[id] = NewDerivedNode(id, m).
				WithTopology(IP).
				WithLatests(map[string]string{docker.ContainerID: containerID}).
				WithCounters(map[string]int{ipsKey: 1})

		}
	}

	return result
}

// MapIP2Container maps IP nodes produced from MapContainer2IP back to
// container nodes.  If there is more than one container with a given
// IP, it is dropped.
func MapIP2Container(n report.Node, _ report.Networks) report.Nodes {
	// If an IP is shared between multiple containers, we can't
	// reliably attribute an connection based on its IP
	if count, _ := n.Counters.Lookup(ipsKey); count > 1 {
		return report.Nodes{}
	}

	// Propagate the internet pseudo node
	if strings.HasSuffix(n.ID, TheInternetID) {
		return report.Nodes{n.ID: n}
	}

	// If this node is not a container, exclude it.
	// This excludes all the nodes we've dragged in from endpoint
	// that we failed to join to a container.
	containerID, ok := n.Latest.Lookup(docker.ContainerID)
	if !ok {
		return report.Nodes{}
	}

	id := report.MakeContainerNodeID(containerID)
	return report.Nodes{
		id: NewDerivedNode(id, n).
			WithTopology(report.Container),
	}
}

// MapEndpoint2Pseudo makes internet of host pesudo nodes from a endpoint node.
func MapEndpoint2Pseudo(n report.Node, local report.Networks) report.Nodes {
	var node report.Node

	addr, ok := n.Latest.Lookup(endpoint.Addr)
	if !ok {
		return report.Nodes{}
	}

	if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
		// If the dstNodeAddr is not in a network local to this report, we emit an
		// internet node
		node = theInternetNode(n)
	} else {
		node = NewDerivedPseudoNode(MakePseudoNodeID(addr), n)
	}
	return report.Nodes{node.ID: node}
}

// MapEndpoint2Process maps endpoint Nodes to process
// Nodes.
//
// If this function is given a pseudo node, then it will just return it;
// Pseudo nodes will never have pids in them, and therefore will never
// be able to be turned into a Process node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a process, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a process graph to get that info.
func MapEndpoint2Process(n report.Node, local report.Networks) report.Nodes {
	// Nodes without a hostid are treated as pseudo nodes
	if _, ok := n.Latest.Lookup(report.HostNodeID); !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	pid, timestamp, ok := n.Latest.LookupEntry(process.PID)
	if !ok {
		return report.Nodes{}
	}

	id := report.MakeProcessNodeID(report.ExtractHostID(n), pid)
	node := NewDerivedNode(id, n).WithTopology(report.Process)
	node.Latest = node.Latest.Set(process.PID, timestamp, pid)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
}

// MapProcess2Container maps process Nodes to container
// Nodes.
//
// If this function is given a node without a docker_container_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapProcess2Container(n report.Node, _ report.Networks) report.Nodes {
	// Propagate the internet pseudo node
	if strings.HasSuffix(n.ID, TheInternetID) {
		return report.Nodes{n.ID: n}
	}

	// Don't propagate non-internet pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{}
	}

	// Otherwise, if the process is not in a container, group it
	// into an per-host "Uncontained" node.  If for whatever reason
	// this node doesn't have a host id in their nodemetadata, it'll
	// all get grouped into a single uncontained node.
	var (
		id   string
		node report.Node
	)
	if containerID, ok := n.Latest.Lookup(docker.ContainerID); ok {
		id = report.MakeContainerNodeID(containerID)
		node = NewDerivedNode(id, n).WithTopology(report.Container)
	} else {
		id = MakePseudoNodeID(UncontainedID, report.ExtractHostID(n))
		node = NewDerivedPseudoNode(id, n)
	}
	return report.Nodes{id: node}
}

// MapProcess2Name maps process Nodes to Nodes
// for each process name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.
func MapProcess2Name(n report.Node, _ report.Networks) report.Nodes {
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	name, timestamp, ok := n.Latest.LookupEntry(process.Name)
	if !ok {
		return report.Nodes{}
	}

	node := NewDerivedNode(name, n).WithTopology(report.Process)
	node.Latest = node.Latest.Set(process.Name, timestamp, name)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{name: node}
}

// MapContainer2ContainerImage maps container Nodes to container
// image Nodes.
//
// If this function is given a node without a docker_image_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapContainer2ContainerImage(n report.Node, _ report.Networks) report.Nodes {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a image_id
	// (maybe slightly out of sync reports), just drop it
	imageID, timestamp, ok := n.Latest.LookupEntry(docker.ImageID)
	if !ok {
		return report.Nodes{}
	}

	// Add container id key to the counters, which will later be counted to produce the minor label
	id := report.MakeContainerImageNodeID(imageID)
	result := NewDerivedNode(id, n).WithTopology(report.ContainerImage)
	result.Latest = result.Latest.Set(docker.ImageID, timestamp, imageID)
	result.Counters = result.Counters.Add(n.Topology, 1)
	return report.Nodes{id: result}
}

// ImageNameWithoutVersion splits the image name apart, returning the name
// without the version, if possible
func ImageNameWithoutVersion(name string) string {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) == 3 {
		name = fmt.Sprintf("%s/%s", parts[1], parts[2])
	}
	parts = strings.SplitN(name, ":", 2)
	return parts[0]
}

// MapX2Host maps any Nodes to host
// Nodes.
//
// If this function is given a node without a hostname
// (including other pseudo nodes), it will drop the node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapX2Host(n report.Node, _ report.Networks) report.Nodes {
	// Don't propagate all pseudo nodes - we do this in MapEndpoint2Host
	if n.Topology == Pseudo {
		return report.Nodes{}
	}
	hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID)
	if !ok {
		return report.Nodes{}
	}
	id := report.MakeHostNodeID(report.ExtractHostID(n))
	result := NewDerivedNode(id, n).WithTopology(report.Host)
	result.Latest = result.Latest.Set(report.HostNodeID, timestamp, hostNodeID)
	result.Counters = result.Counters.Add(n.Topology, 1)
	return report.Nodes{id: result}
}

// MapEndpoint2Host takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
func MapEndpoint2Host(n report.Node, local report.Networks) report.Nodes {
	// Nodes without a hostid are treated as pseudo nodes
	hostNodeID, timestamp, ok := n.Latest.LookupEntry(report.HostNodeID)
	if !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	id := report.MakeHostNodeID(report.ExtractHostID(n))
	result := NewDerivedNode(id, n).WithTopology(report.Host)
	result.Latest = result.Latest.Set(report.HostNodeID, timestamp, hostNodeID)
	result.Counters = result.Counters.Add(n.Topology, 1)
	return report.Nodes{id: result}
}

// MapContainer2Pod maps container Nodes to pod
// Nodes.
//
// If this function is given a node without a kubernetes_pod_id
// (including other pseudo nodes), it will produce an "Unmanaged"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapContainer2Pod(n report.Node, _ report.Networks) report.Nodes {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a pod_id (maybe
	// slightly out of sync reports, or its not in a pod), just drop it
	namespace, ok := n.Latest.Lookup(kubernetes.Namespace)
	if !ok {
		return report.Nodes{}
	}
	podID, ok := n.Latest.Lookup(kubernetes.PodID)
	if !ok {
		return report.Nodes{}
	}
	podName := strings.TrimPrefix(podID, namespace+"/")
	id := report.MakePodNodeID(namespace, podName)

	// Due to a bug in kubernetes, addon pods on the master node are not returned
	// from the API. Adding the namespace and pod name is a workaround until
	// https://github.com/kubernetes/kubernetes/issues/14738 is fixed.
	return report.Nodes{
		id: NewDerivedNode(id, n).
			WithTopology(report.Pod).
			WithLatests(map[string]string{
				kubernetes.Namespace: namespace,
				kubernetes.PodName:   podName,
			}),
	}
}

// MapPod2Service maps pod Nodes to service Nodes.
//
// If this function is given a node without a kubernetes_pod_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a pod graph to get that info.
func MapPod2Service(pod report.Node, _ report.Networks) report.Nodes {
	// Propagate all pseudo nodes
	if pod.Topology == Pseudo {
		return report.Nodes{pod.ID: pod}
	}

	// Otherwise, if some some reason the pod doesn't have a service_ids (maybe
	// slightly out of sync reports, or its not in a service), just drop it
	namespace, ok := pod.Latest.Lookup(kubernetes.Namespace)
	if !ok {
		return report.Nodes{}
	}
	ids, ok := pod.Latest.Lookup(kubernetes.ServiceIDs)
	if !ok {
		return report.Nodes{}
	}

	result := report.Nodes{}
	for _, serviceID := range strings.Fields(ids) {
		serviceName := strings.TrimPrefix(serviceID, namespace+"/")
		id := report.MakeServiceNodeID(namespace, serviceName)
		result[id] = NewDerivedNode(id, pod).WithTopology(report.Service)
	}
	return result
}

// MapContainer2Hostname maps container Nodes to 'hostname' renderabled nodes..
func MapContainer2Hostname(n report.Node, _ report.Networks) report.Nodes {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a hostname
	// (maybe slightly out of sync reports), just drop it
	id, timestamp, ok := n.Latest.LookupEntry(docker.ContainerHostname)
	if !ok {
		return report.Nodes{}
	}

	node := NewDerivedNode(id, n)
	node.Latest = node.Latest.
		Set(docker.ContainerHostname, timestamp, id).
		Delete(docker.ContainerName) // TODO(paulbellamy): total hack to render these by hostname instead.
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
}
