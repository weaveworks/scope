package render

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	UncontainedID    = "uncontained"
	UncontainedMajor = "Uncontained"

	TheInternetID    = "theinternet"
	TheInternetMajor = "The Internet"

	containersKey = "containers"
	podsKey       = "pods"
	processesKey  = "processes"
	servicesKey   = "services"

	AmazonECSContainerNameLabel = "com.amazonaws.ecs.container-name"
)

// MapFunc is anything which can take an arbitrary RenderableNode and
// return a set of other RenderableNodes.
//
// If the output is empty, the node shall be omitted from the rendered topology.
type MapFunc func(RenderableNode, report.Networks) RenderableNodes

// MapEndpointIdentity maps an endpoint topology node to a single endpoint
// renderable node. As it is only ever run on endpoint topology nodes, we
// expect that certain keys are present.
func MapEndpointIdentity(m RenderableNode, local report.Networks) RenderableNodes {
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNodes{}
	}

	port, ok := m.Metadata[endpoint.Port]
	if !ok {
		return RenderableNodes{}
	}

	// We only show nodes found through procspy in this view.
	_, procspied := m.Metadata[endpoint.Procspied]
	if !procspied {
		return RenderableNodes{}
	}

	// Nodes without a hostid are treated as psuedo nodes
	if _, ok = m.Metadata[report.HostNodeID]; !ok {
		// If the dstNodeAddr is not in a network local to this report, we emit an
		// internet node
		if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
			return RenderableNodes{TheInternetID: newDerivedPseudoNode(TheInternetID, TheInternetMajor, m)}
		}

		// We are a 'client' pseudo node if the port is in the ephemeral port range.
		// Linux uses 32768 to 61000, IANA suggests 49152 to 65535.
		if p, err := strconv.Atoi(port); err == nil && len(m.Adjacency) > 0 && p >= 32768 && p < 65535 {
			// We only exist if there is something in our adjacency
			// Generate a single pseudo node for every (client ip, server ip, server port)
			dstNodeID := m.Adjacency[0]
			serverIP, serverPort := trySplitAddr(dstNodeID)
			outputID := MakePseudoNodeID(addr, serverIP, serverPort)
			return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr, m)}
		}

		// Otherwise (the server node is missing), generate a pseudo node for every (server ip, server port)
		outputID := MakePseudoNodeID(addr, port)
		if port != "" {
			return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr+":"+port, m)}
		}
		return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr, m)}
	}

	var (
		id    = MakeEndpointID(report.ExtractHostID(m.Node), addr, port)
		major = fmt.Sprintf("%s:%s", addr, port)
		minor = report.ExtractHostID(m.Node)
		rank  = major
	)

	pid, pidOK := m.Metadata[process.PID]
	if pidOK {
		minor = fmt.Sprintf("%s (%s)", minor, pid)
	}

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapProcessIdentity maps a process topology node to a process renderable
// node. As it is only ever run on process topology nodes, we expect that
// certain keys are present.
func MapProcessIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	pid, ok := m.Metadata[process.PID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		id    = MakeProcessID(report.ExtractHostID(m.Node), pid)
		major = m.Metadata["comm"]
		minor = fmt.Sprintf("%s (%s)", report.ExtractHostID(m.Node), pid)
		rank  = m.Metadata["comm"]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapContainerIdentity maps a container topology node to a container
// renderable node. As it is only ever run on container topology nodes, we
// expect that certain keys are present.
func MapContainerIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	id, ok := m.Metadata[docker.ContainerID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		major, _ = GetRenderableContainerName(m.Node)
		minor    = report.ExtractHostID(m.Node)
		rank     = m.Metadata[docker.ImageID]
	)

	node := NewRenderableNodeWith(id, major, minor, rank, m)
	if imageID, ok := m.Metadata[docker.ImageID]; ok {
		hostID, _, _ := report.ParseContainerNodeID(m.ID)
		node.Origins = node.Origins.Add(report.MakeContainerNodeID(hostID, imageID))
	}
	return RenderableNodes{id: node}
}

// GetRenderableContainerName obtains a user-friendly container name, to render in the UI
func GetRenderableContainerName(nmd report.Node) (string, bool) {
	// Amazon's ecs-agent produces huge Docker container names, destructively
	// derived from mangling Container Definition names in Task
	// Definitions.
	//
	// However, the ecs-agent provides a label containing the original Container
	// Definition name.
	if labelValue, ok := nmd.Metadata[docker.LabelPrefix+AmazonECSContainerNameLabel]; ok {
		return labelValue, true
	}

	name, ok := nmd.Metadata[docker.ContainerName]
	return name, ok
}

// MapContainerImageIdentity maps a container image topology node to container
// image renderable node. As it is only ever run on container image topology
// nodes, we expect that certain keys are present.
func MapContainerImageIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	id, ok := m.Metadata[docker.ImageID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		major = m.Metadata[docker.ImageName]
		rank  = m.Metadata[docker.ImageID]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, "", rank, m)}
}

// MapPodIdentity maps a pod topology node to pod renderable node. As it is
// only ever run on pod topology nodes, we expect that certain keys
// are present.
func MapPodIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	id, ok := m.Metadata[kubernetes.PodID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		major = m.Metadata[kubernetes.PodName]
		rank  = m.Metadata[kubernetes.PodID]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, "", rank, m)}
}

// MapServiceIdentity maps a service topology node to service renderable node. As it is
// only ever run on service topology nodes, we expect that certain keys
// are present.
func MapServiceIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	id, ok := m.Metadata[kubernetes.ServiceID]
	if !ok {
		return RenderableNodes{}
	}

	var (
		major = m.Metadata[kubernetes.ServiceName]
		rank  = m.Metadata[kubernetes.ServiceID]
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, "", rank, m)}
}

// MapAddressIdentity maps an address topology node to an address renderable
// node. As it is only ever run on address topology nodes, we expect that
// certain keys are present.
func MapAddressIdentity(m RenderableNode, local report.Networks) RenderableNodes {
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNodes{}
	}

	// Conntracked connections don't have a host id unless
	// they were merged with a procspied connection.  Filter
	// out those that weren't.
	_, hasHostID := m.Metadata[report.HostNodeID]
	_, conntracked := m.Metadata[endpoint.Conntracked]
	if !hasHostID && conntracked {
		return RenderableNodes{}
	}

	// Nodes without a hostid are treated as psuedo nodes
	if !hasHostID {
		// If the addr is not in a network local to this report, we emit an
		// internet node
		if !local.Contains(net.ParseIP(addr)) {
			return RenderableNodes{TheInternetID: newDerivedPseudoNode(TheInternetID, TheInternetMajor, m)}
		}

		// Otherwise generate a pseudo node for every
		outputID := MakePseudoNodeID(addr, "")
		if len(m.Adjacency) > 0 {
			_, dstAddr, _ := report.ParseAddressNodeID(m.Adjacency[0])
			outputID = MakePseudoNodeID(addr, dstAddr)
		}
		return RenderableNodes{outputID: newDerivedPseudoNode(outputID, addr, m)}
	}

	var (
		id    = MakeAddressID(report.ExtractHostID(m.Node), addr)
		major = addr
		minor = report.ExtractHostID(m.Node)
		rank  = major
	)

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapHostIdentity maps a host topology node to a host renderable node. As it
// is only ever run on host topology nodes, we expect that certain keys are
// present.
func MapHostIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	var (
		id                 = MakeHostID(report.ExtractHostID(m.Node))
		hostname           = m.Metadata[host.HostName]
		parts              = strings.SplitN(hostname, ".", 2)
		major, minor, rank = "", "", ""
	)

	if len(parts) == 2 {
		major, minor, rank = parts[0], parts[1], parts[1]
	} else {
		major = hostname
	}

	return RenderableNodes{id: NewRenderableNodeWith(id, major, minor, rank, m)}
}

// MapEndpoint2IP maps endpoint nodes to their IP address, for joining
// with container nodes.  We drop endpoint nodes with pids, as they
// will be joined to containers through the process topology, and we
// don't want to double count edges.
func MapEndpoint2IP(m RenderableNode, local report.Networks) RenderableNodes {
	// Don't include procspied connections, to prevent double counting
	_, ok := m.Metadata[endpoint.Procspied]
	if ok {
		return RenderableNodes{}
	}
	addr, ok := m.Metadata[endpoint.Addr]
	if !ok {
		return RenderableNodes{}
	}
	if !local.Contains(net.ParseIP(addr)) {
		return RenderableNodes{TheInternetID: newDerivedPseudoNode(TheInternetID, TheInternetMajor, m)}
	}

	result := RenderableNodes{addr: NewRenderableNodeWith(addr, "", "", "", m)}
	// Emit addr:port nodes as well, so connections from the internet to containers
	// via port mapping also works.
	port, ok := m.Metadata[endpoint.Port]
	if ok {
		id := fmt.Sprintf("%s:%s", addr, port)
		result[id] = NewRenderableNodeWith(id, "", "", "", m)
	}
	return result
}

var portMappingMatch = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}):([0-9]+)->([0-9]+)/tcp`)

// MapContainer2IP maps container nodes to their IP addresses (outputs
// multiple nodes).  This allows container to be joined directly with
// the endpoint topology.
func MapContainer2IP(m RenderableNode, _ report.Networks) RenderableNodes {
	result := RenderableNodes{}
	addrs, ok := m.Metadata[docker.ContainerIPs]
	if !ok {
		return result
	}
	for _, addr := range strings.Fields(addrs) {
		node := NewRenderableNodeWith(addr, "", "", "", m)
		node.Counters[containersKey] = 1
		result[addr] = node
	}

	// also output all the host:port port mappings
	for _, mapping := range portMappingMatch.FindAllStringSubmatch(m.Metadata[docker.ContainerPorts], -1) {
		ip, port := mapping[1], mapping[2]
		id := fmt.Sprintf("%s:%s", ip, port)
		node := NewRenderableNodeWith(id, "", "", "", m)
		node.Counters[containersKey] = 1
		result[id] = node
	}

	return result
}

// MapIP2Container maps IP nodes produced from MapContainer2IP back to
// container nodes.  If there is more than one container with a given
// IP, it is dropped.
func MapIP2Container(n RenderableNode, _ report.Networks) RenderableNodes {
	// If an IP is shared between multiple containers, we can't
	// reliably attribute an connection based on its IP
	if n.Node.Counters[containersKey] > 1 {
		return RenderableNodes{}
	}

	// Propogate the internet pseudo node.
	if n.ID == TheInternetID {
		return RenderableNodes{n.ID: n}
	}

	// If this node is not a container, exclude it.
	// This excludes all the nodes we've dragged in from endpoint
	// that we failed to join to a container.
	id, ok := n.Node.Metadata[docker.ContainerID]
	if !ok {
		return RenderableNodes{}
	}

	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// MapEndpoint2Process maps endpoint RenderableNodes to process
// RenderableNodes.
//
// If this function is given a pseudo node, then it will just return it;
// Pseudo nodes will never have pids in them, and therefore will never
// be able to be turned into a Process node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a process, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a process graph to get that info.
func MapEndpoint2Process(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	pid, ok := n.Node.Metadata[process.PID]
	if !ok {
		return RenderableNodes{}
	}

	id := MakeProcessID(report.ExtractHostID(n.Node), pid)
	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// MapProcess2Container maps process RenderableNodes to container
// RenderableNodes.
//
// If this function is given a node without a docker_container_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapProcess2Container(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate the internet pseudo node
	if n.ID == TheInternetID {
		return RenderableNodes{n.ID: n}
	}

	// Don't propogate non-internet pseudo nodes
	if n.Pseudo {
		return RenderableNodes{}
	}

	// Otherwise, if the process is not in a container, group it
	// into an per-host "Uncontained" node.  If for whatever reason
	// this node doesn't have a host id in their nodemetadata, it'll
	// all get grouped into a single uncontained node.
	id, ok := n.Node.Metadata[docker.ContainerID]
	if !ok {
		hostID := report.ExtractHostID(n.Node)
		id = MakePseudoNodeID(UncontainedID, hostID)
		node := newDerivedPseudoNode(id, UncontainedMajor, n)
		node.LabelMinor = hostID
		return RenderableNodes{id: node}
	}

	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// MapProcess2Name maps process RenderableNodes to RenderableNodes
// for each process name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.  Therefore
// it outputs a properly-formed node with labels etc.
func MapProcess2Name(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	name, ok := n.Node.Metadata["comm"]
	if !ok {
		return RenderableNodes{}
	}

	node := NewDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	node.Node.Counters[processesKey] = 1
	return RenderableNodes{name: node}
}

// MapCountProcessName maps 1:1 process name nodes, counting
// the number of processes grouped together and putting
// that info in the minor label.
func MapCountProcessName(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	processes := n.Node.Counters[processesKey]
	if processes == 1 {
		n.LabelMinor = "1 process"
	} else {
		n.LabelMinor = fmt.Sprintf("%d processes", processes)
	}
	return RenderableNodes{n.ID: n}
}

// MapContainer2ContainerImage maps container RenderableNodes to container
// image RenderableNodes.
//
// If this function is given a node without a docker_image_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapContainer2ContainerImage(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a image_id
	// (maybe slightly out of sync reports), just drop it
	id, ok := n.Node.Metadata[docker.ImageID]
	if !ok {
		return RenderableNodes{}
	}

	// Add container id key to the counters, which will later be counted to produce the minor label
	result := NewDerivedNode(id, n)
	result.Node.Counters[containersKey] = 1
	return RenderableNodes{id: result}
}

// MapPod2Service maps pod RenderableNodes to service RenderableNodes.
//
// If this function is given a node without a kubernetes_pod_id
// (including other pseudo nodes), it will produce an "Uncontained"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a pod graph to get that info.
func MapPod2Service(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the pod doesn't have a service_ids (maybe
	// slightly out of sync reports, or its not in a service), just drop it
	ids, ok := n.Node.Metadata[kubernetes.ServiceIDs]
	if !ok {
		return RenderableNodes{}
	}

	result := RenderableNodes{}
	for _, id := range strings.Fields(ids) {
		n := NewDerivedNode(id, n)
		n.Node.Counters[podsKey] = 1
		result[id] = n
	}
	return result
}

func imageNameWithoutVersion(name string) string {
	parts := strings.SplitN(name, ":", 2)
	if len(parts) == 2 {
		return parts[0]
	}
	return name
}

// MapContainerImage2Name maps container images RenderableNodes to
// RenderableNodes for each container image name.
//
// This mapper is unlike the other foo2bar mappers as the intention
// is not to join the information with another topology.  Therefore
// it outputs a properly-formed node with labels etc.
func MapContainerImage2Name(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	name, ok := n.Node.Metadata[docker.ImageName]
	if !ok {
		return RenderableNodes{}
	}

	name = imageNameWithoutVersion(name)

	node := NewDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	node.Node = n.Node.Copy() // Propagate NMD for container counting.
	return RenderableNodes{name: node}
}

// MapContainer2Pod maps container RenderableNodes to pod
// RenderableNodes.
//
// If this function is given a node without a kubernetes_pod_id
// (including other pseudo nodes), it will produce an "Unmanaged"
// pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapContainer2Pod(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a pod_id (maybe
	// slightly out of sync reports, or its not in a pod), just drop it
	id, ok := n.Node.Metadata["docker_label_io.kubernetes.pod.name"]
	if !ok {
		return RenderableNodes{}
	}

	// Add container-<id> key to NMD, which will later be counted to produce the
	// minor label
	result := NewRenderableNodeWith(id, "", "", id, n)
	result.Node.Counters[containersKey] = 1
	// Due to a bug in kubernetes, addon pods on the master node are not returned
	// from the API. This is a workaround until
	// https://github.com/kubernetes/kubernetes/issues/14738 is fixed.
	if s := strings.SplitN(id, "/", 2); len(s) == 2 {
		result.LabelMajor = s[1]
		result.Node.Metadata[kubernetes.Namespace] = s[0]
		result.Node.Metadata[kubernetes.PodName] = s[1]
	}
	return RenderableNodes{id: result}
}

// MapContainer2Hostname maps container RenderableNodes to 'hostname' renderabled nodes..
func MapContainer2Hostname(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propogate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a hostname
	// (maybe slightly out of sync reports), just drop it
	id, ok := n.Node.Metadata[docker.ContainerHostname]
	if !ok {
		return RenderableNodes{}
	}

	result := NewDerivedNode(id, n)
	result.LabelMajor = id
	result.Rank = id

	// Add container id key to the counters, which will later be counted to produce the minor label
	result.Node.Counters[containersKey] = 1
	return RenderableNodes{id: result}
}

// MapCountContainers maps 1:1 container image nodes, counting
// the number of containers grouped together and putting
// that info in the minor label.
func MapCountContainers(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	containers := n.Node.Counters[containersKey]
	if containers == 1 {
		n.LabelMinor = "1 container"
	} else {
		n.LabelMinor = fmt.Sprintf("%d containers", containers)
	}
	return RenderableNodes{n.ID: n}
}

// MapCountPods maps 1:1 service nodes, counting the number of pods grouped
// together and putting that info in the minor label.
func MapCountPods(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	pods := n.Node.Counters[podsKey]
	if pods == 1 {
		n.LabelMinor = "1 pod"
	} else {
		n.LabelMinor = fmt.Sprintf("%d pods", pods)
	}
	return RenderableNodes{n.ID: n}
}

// MapAddress2Host maps address RenderableNodes to host RenderableNodes.
//
// Otherthan pseudo nodes, we can assume all nodes have a HostID
func MapAddress2Host(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	id := MakeHostID(report.ExtractHostID(n.Node))
	return RenderableNodes{id: NewDerivedNode(id, n)}
}

// trySplitAddr is basically ParseArbitraryNodeID, since its callsites
// (pseudo funcs) just have opaque node IDs and don't know what topology they
// come from. Without changing how pseudo funcs work, we can't make it much
// smarter.
//
// TODO change how pseudofuncs work, and eliminate this helper.
func trySplitAddr(addr string) (string, string) {
	fields := strings.SplitN(addr, report.ScopeDelim, 3)
	if len(fields) == 3 {
		return fields[1], fields[2]
	}
	if len(fields) == 2 {
		return fields[1], ""
	}
	panic(addr)
}
