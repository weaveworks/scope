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

	TheInternetID      = "theinternet"
	IncomingInternetID = "in-" + TheInternetID
	OutgoingInternetID = "out-" + TheInternetID
	InboundMajor       = "The Internet"
	OutboundMajor      = "The Internet"
	InboundMinor       = "Inbound connections"
	OutboundMinor      = "Outbound connections"

	ContainersKey = "containers"
	ipsKey        = "ips"
	podsKey       = "pods"
	processesKey  = "processes"
	servicesKey   = "services"

	AmazonECSContainerNameLabel  = "com.amazonaws.ecs.container-name"
	KubernetesContainerNameLabel = "io.kubernetes.container.name"
)

// MapFunc is anything which can take an arbitrary RenderableNode and
// return a set of other RenderableNodes.
//
// If the output is empty, the node shall be omitted from the rendered topology.
type MapFunc func(RenderableNode, report.Networks) RenderableNodes

func theInternetNode(m RenderableNode) RenderableNode {
	node := newDerivedPseudoNode("", "", m)
	node.Shape = Cloud
	// emit one internet node for incoming, one for outgoing
	if len(m.Adjacency) > 0 {
		node.ID = IncomingInternetID
		node.LabelMajor = InboundMajor
		node.LabelMinor = InboundMinor
	} else {
		node.ID = OutgoingInternetID
		node.LabelMajor = OutboundMajor
		node.LabelMinor = OutboundMinor
	}
	return node
}

// MapEndpointIdentity remaps endpoints to have an id format consistent
// with render/id.go; no pseudo nodes are introduced in this step, so
// that pseudo nodes introduces later are guaranteed to have endpoints
// as children.  This is needed to construct the connection details tables.
func MapEndpointIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	addr, ok := m.Latest.Lookup(endpoint.Addr)
	if !ok {
		return RenderableNodes{}
	}

	port, ok := m.Latest.Lookup(endpoint.Port)
	if !ok {
		return RenderableNodes{}
	}

	// We only show nodes found through procspy in this view.
	_, procspied := m.Latest.Lookup(endpoint.Procspied)
	if !procspied {
		return RenderableNodes{}
	}

	id := MakeEndpointID(report.ExtractHostID(m.Node), addr, port)
	return RenderableNodes{id: NewRenderableNodeWith(id, "", "", "", m)}
}

// MapProcessIdentity maps a process topology node to a process renderable
// node. As it is only ever run on process topology nodes, we expect that
// certain keys are present.
func MapProcessIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	pid, ok := m.Latest.Lookup(process.PID)
	if !ok {
		return RenderableNodes{}
	}

	var (
		id       = MakeProcessID(report.ExtractHostID(m.Node), pid)
		major, _ = m.Latest.Lookup(process.Name)
		minor    = fmt.Sprintf("%s (%s)", report.ExtractHostID(m.Node), pid)
		rank, _  = m.Latest.Lookup(process.Name)
	)

	node := NewRenderableNodeWith(id, major, minor, rank, m)
	node.Shape = Square
	return RenderableNodes{id: node}
}

// MapContainerIdentity maps a container topology node to a container
// renderable node. As it is only ever run on container topology nodes, we
// expect that certain keys are present.
func MapContainerIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	containerID, ok := m.Latest.Lookup(docker.ContainerID)
	if !ok {
		return RenderableNodes{}
	}

	var (
		id       = MakeContainerID(containerID)
		major, _ = GetRenderableContainerName(m.Node)
		minor    = report.ExtractHostID(m.Node)
	)

	node := NewRenderableNodeWith(id, major, minor, "", m)
	node.ControlNode = m.ID
	node.Shape = Hexagon
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
	if labelValue, ok := nmd.Latest.Lookup(docker.LabelPrefix + AmazonECSContainerNameLabel); ok {
		return labelValue, true
	}

	// Kubernetes also mangles its Docker container names and provides a
	// label with the original container name. However, note that this label
	// is only provided by Kubernetes versions >= 1.2 (see
	// https://github.com/kubernetes/kubernetes/pull/17234/ )
	if labelValue, ok := nmd.Latest.Lookup(docker.LabelPrefix + KubernetesContainerNameLabel); ok {
		return labelValue, true
	}

	name, ok := nmd.Latest.Lookup(docker.ContainerName)
	return name, ok
}

// MapContainerImageIdentity maps a container image topology node to container
// image renderable node. As it is only ever run on container image topology
// nodes, we expect that certain keys are present.
func MapContainerImageIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	imageID, ok := m.Latest.Lookup(docker.ImageID)
	if !ok {
		return RenderableNodes{}
	}

	var (
		id       = MakeContainerImageID(imageID)
		major, _ = m.Latest.Lookup(docker.ImageName)
		rank     = imageID
	)

	node := NewRenderableNodeWith(id, major, "", rank, m)
	node.Shape = Hexagon
	node.Stack = true
	return RenderableNodes{id: node}
}

// MapPodIdentity maps a pod topology node to pod renderable node. As it is
// only ever run on pod topology nodes, we expect that certain keys
// are present.
func MapPodIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	podID, ok := m.Latest.Lookup(kubernetes.PodID)
	if !ok {
		return RenderableNodes{}
	}

	var (
		id       = MakePodID(podID)
		major, _ = m.Latest.Lookup(kubernetes.PodName)
		rank, _  = m.Latest.Lookup(kubernetes.PodID)
	)

	node := NewRenderableNodeWith(id, major, "", rank, m)
	node.Shape = Heptagon
	return RenderableNodes{id: node}
}

// MapServiceIdentity maps a service topology node to service renderable node. As it is
// only ever run on service topology nodes, we expect that certain keys
// are present.
func MapServiceIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	serviceID, ok := m.Latest.Lookup(kubernetes.ServiceID)
	if !ok {
		return RenderableNodes{}
	}

	var (
		id       = MakeServiceID(serviceID)
		major, _ = m.Latest.Lookup(kubernetes.ServiceName)
		rank, _  = m.Latest.Lookup(kubernetes.ServiceID)
	)

	node := NewRenderableNodeWith(id, major, "", rank, m)
	node.Shape = Heptagon
	node.Stack = true
	return RenderableNodes{id: node}
}

// MapHostIdentity maps a host topology node to a host renderable node. As it
// is only ever run on host topology nodes, we expect that certain keys are
// present.
func MapHostIdentity(m RenderableNode, _ report.Networks) RenderableNodes {
	var (
		id                 = MakeHostID(report.ExtractHostID(m.Node))
		hostname, _        = m.Latest.Lookup(host.HostName)
		parts              = strings.SplitN(hostname, ".", 2)
		major, minor, rank = "", "", ""
	)

	if len(parts) == 2 {
		major, minor, rank = parts[0], parts[1], parts[1]
	} else {
		major = hostname
	}

	node := NewRenderableNodeWith(id, major, minor, rank, m)
	node.Shape = Circle
	return RenderableNodes{id: node}
}

// MapEndpoint2IP maps endpoint nodes to their IP address, for joining
// with container nodes.  We drop endpoint nodes with pids, as they
// will be joined to containers through the process topology, and we
// don't want to double count edges.
func MapEndpoint2IP(m RenderableNode, local report.Networks) RenderableNodes {
	// Don't include procspied connections, to prevent double counting
	_, ok := m.Latest.Lookup(endpoint.Procspied)
	if ok {
		return RenderableNodes{}
	}
	scope, addr, port, ok := report.ParseEndpointNodeID(m.ID)
	if !ok {
		return RenderableNodes{}
	}
	if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
		return RenderableNodes{TheInternetID: theInternetNode(m)}
	}

	// We don't always know what port a container is listening on, and
	// container-to-container communications can be unambiguously identified
	// without ports. OTOH, connections to the host IPs which have been port
	// mapped to a container can only be unambiguously identified with the port.
	// So we need to emit two nodes, for two different cases.
	id := report.MakeScopedEndpointNodeID(scope, addr, "")
	idWithPort := report.MakeScopedEndpointNodeID(scope, addr, port)
	m = m.WithParents(report.EmptySets)
	return RenderableNodes{
		id:         NewRenderableNodeWith(id, "", "", "", m),
		idWithPort: NewRenderableNodeWith(idWithPort, "", "", "", m),
	}
}

var portMappingMatch = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}):([0-9]+)->([0-9]+)/tcp`)

// MapContainer2IP maps container nodes to their IP addresses (outputs
// multiple nodes).  This allows container to be joined directly with
// the endpoint topology.
func MapContainer2IP(m RenderableNode, _ report.Networks) RenderableNodes {
	result := RenderableNodes{}
	if addrs, ok := m.Sets.Lookup(docker.ContainerIPsWithScopes); ok {
		for _, addr := range addrs {
			scope, addr, ok := report.ParseAddressNodeID(addr)
			if !ok {
				continue
			}
			id := report.MakeScopedEndpointNodeID(scope, addr, "")
			node := NewRenderableNodeWith(id, "", "", "", m)
			node.Counters = node.Counters.Add(ipsKey, 1)
			result[id] = node
		}
	}

	// Also output all the host:port port mappings (see above comment).
	// In this case we assume this doesn't need a scope, as they are for host IPs.
	ports, _ := m.Sets.Lookup(docker.ContainerPorts)
	for _, portMapping := range ports {
		if mapping := portMappingMatch.FindStringSubmatch(portMapping); mapping != nil {
			ip, port := mapping[1], mapping[2]
			id := report.MakeScopedEndpointNodeID("", ip, port)
			node := NewRenderableNodeWith(id, "", "", "", m.WithParents(report.EmptySets))
			node.Counters = node.Counters.Add(ipsKey, 1)
			result[id] = node
		}
	}

	return result
}

// MapIP2Container maps IP nodes produced from MapContainer2IP back to
// container nodes.  If there is more than one container with a given
// IP, it is dropped.
func MapIP2Container(n RenderableNode, _ report.Networks) RenderableNodes {
	// If an IP is shared between multiple containers, we can't
	// reliably attribute an connection based on its IP
	if count, _ := n.Node.Counters.Lookup(ipsKey); count > 1 {
		return RenderableNodes{}
	}

	// Propagate the internet pseudo node
	if strings.HasSuffix(n.ID, TheInternetID) {
		return RenderableNodes{n.ID: n}
	}

	// If this node is not a container, exclude it.
	// This excludes all the nodes we've dragged in from endpoint
	// that we failed to join to a container.
	containerID, ok := n.Node.Latest.Lookup(docker.ContainerID)
	if !ok {
		return RenderableNodes{}
	}

	id := MakeContainerID(containerID)
	node := NewDerivedNode(id, n.WithParents(report.EmptySets))
	node.Shape = Hexagon
	return RenderableNodes{id: node}
}

// MapEndpoint2Pseudo makes internet of host pesudo nodes from a endpoint node.
func MapEndpoint2Pseudo(n RenderableNode, local report.Networks) RenderableNodes {
	var node RenderableNode

	addr, ok := n.Latest.Lookup(endpoint.Addr)
	if !ok {
		return RenderableNodes{}
	}

	port, ok := n.Latest.Lookup(endpoint.Port)
	if !ok {
		return RenderableNodes{}
	}

	if ip := net.ParseIP(addr); ip != nil && !local.Contains(ip) {
		// If the dstNodeAddr is not in a network local to this report, we emit an
		// internet node
		node = theInternetNode(n)

	} else if p, err := strconv.Atoi(port); err == nil && len(n.Adjacency) > 0 && p >= 32768 && p < 65535 {
		// We are a 'client' pseudo node if the port is in the ephemeral port range.
		// Linux uses 32768 to 61000, IANA suggests 49152 to 65535.
		// We only exist if there is something in our adjacency
		// Generate a single pseudo node for every (client ip, server ip, server port)
		_, serverIP, serverPort, _ := ParseEndpointID(n.Adjacency[0])
		node = newDerivedPseudoNode(MakePseudoNodeID(addr, serverIP, serverPort), addr, n)

	} else if port != "" {
		// Otherwise (the server node is missing), generate a pseudo node for every (server ip, server port)
		node = newDerivedPseudoNode(MakePseudoNodeID(addr, port), addr+":"+port, n)

	} else {
		// Empty port for some reason...
		node = newDerivedPseudoNode(MakePseudoNodeID(addr, port), addr, n)
	}

	node.Children = node.Children.Add(n)
	return RenderableNodes{node.ID: node}
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
func MapEndpoint2Process(n RenderableNode, local report.Networks) RenderableNodes {
	// Nodes without a hostid are treated as pseudo nodes
	if _, ok := n.Latest.Lookup(report.HostNodeID); !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	pid, ok := n.Node.Latest.Lookup(process.PID)
	if !ok {
		return RenderableNodes{}
	}

	id := MakeProcessID(report.ExtractHostID(n.Node), pid)
	node := NewDerivedNode(id, n.WithParents(report.EmptySets))
	node.Shape = Square
	node.Children = node.Children.Add(n)
	return RenderableNodes{id: node}
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
	// Propagate the internet pseudo node
	if strings.HasSuffix(n.ID, TheInternetID) {
		return RenderableNodes{n.ID: n}
	}

	// Don't propagate non-internet pseudo nodes
	if n.Pseudo {
		return RenderableNodes{}
	}

	// Otherwise, if the process is not in a container, group it
	// into an per-host "Uncontained" node.  If for whatever reason
	// this node doesn't have a host id in their nodemetadata, it'll
	// all get grouped into a single uncontained node.
	var (
		id     string
		node   RenderableNode
		hostID = report.ExtractHostID(n.Node)
	)
	n = n.WithParents(report.EmptySets)
	if containerID, ok := n.Node.Latest.Lookup(docker.ContainerID); ok {
		id = MakeContainerID(containerID)
		node = NewDerivedNode(id, n)
		node.Shape = Hexagon
	} else {
		nCopy := n.Copy()
		nCopy.Node = nCopy.Node.WithID("").WithTopology("") // Wipe the ID so it cannot be rendered.
		id = MakePseudoNodeID(UncontainedID, hostID)
		node = newDerivedPseudoNode(id, UncontainedMajor, nCopy)
		node.LabelMinor = hostID
		node.Shape = Square
		node.Stack = true
	}

	node.Children = node.Children.Add(n)
	return RenderableNodes{id: node}
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

	name, ok := n.Node.Latest.Lookup(process.Name)
	if !ok {
		return RenderableNodes{}
	}

	node := NewDerivedNode(name, n)
	node.LabelMajor = name
	node.Rank = name
	node.Counters = node.Node.Counters.Add(processesKey, 1)
	node.Node.Topology = "process_name"
	node.Node.ID = name
	node.Children = node.Children.Add(n)
	node.Shape = Square
	node.Stack = true
	return RenderableNodes{name: node}
}

// MapCountProcessName maps 1:1 process name nodes, counting
// the number of processes grouped together and putting
// that info in the minor label.
func MapCountProcessName(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	output := n.Copy()
	processes, _ := n.Node.Counters.Lookup(processesKey)
	if processes == 1 {
		output.LabelMinor = "1 process"
	} else {
		output.LabelMinor = fmt.Sprintf("%d processes", processes)
	}
	return RenderableNodes{output.ID: output}
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
	// Propagate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a image_id
	// (maybe slightly out of sync reports), just drop it
	imageID, ok := n.Node.Latest.Lookup(docker.ImageID)
	if !ok {
		return RenderableNodes{}
	}

	// Add container id key to the counters, which will later be counted to produce the minor label
	id := MakeContainerImageID(imageID)
	result := NewDerivedNode(id, n.WithParents(report.EmptySets))
	result.Node.Counters = result.Node.Counters.Add(ContainersKey, 1)

	// Add the container as a child of the new image node
	result.Children = result.Children.Add(n)

	result.Node.Topology = "container_image"
	result.Node.ID = report.MakeContainerImageNodeID(imageID)
	result.Shape = Hexagon
	result.Stack = true
	return RenderableNodes{id: result}
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

	name, ok := n.Node.Latest.Lookup(docker.ImageName)
	if !ok {
		return RenderableNodes{}
	}

	name = ImageNameWithoutVersion(name)
	id := MakeContainerImageID(name)

	node := NewDerivedNode(id, n)
	node.LabelMajor = name
	node.Rank = name
	node.Node = n.Node.Copy() // Propagate NMD for container counting.
	node.Shape = Hexagon
	node.Stack = true
	return RenderableNodes{id: node}
}

// MapX2Host maps any RenderableNodes to host
// RenderableNodes.
//
// If this function is given a node without a hostname
// (including other pseudo nodes), it will drop the node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapX2Host(n RenderableNode, _ report.Networks) RenderableNodes {
	// Don't propagate all pseudo nodes - we do this in MapEndpoint2Host
	if n.Pseudo {
		return RenderableNodes{}
	}
	if _, ok := n.Node.Latest.Lookup(report.HostNodeID); !ok {
		return RenderableNodes{}
	}
	id := MakeHostID(report.ExtractHostID(n.Node))
	result := NewDerivedNode(id, n.WithParents(report.EmptySets))
	result.Children = result.Children.Add(n)
	result.Shape = Circle
	result.EdgeMetadata = report.EdgeMetadata{} // Don't double count edge metadata
	return RenderableNodes{id: result}
}

// MapEndpoint2Host takes nodes from the endpoint topology and produces
// host nodes or pseudo nodes.
func MapEndpoint2Host(n RenderableNode, local report.Networks) RenderableNodes {
	// Nodes without a hostid are treated as pseudo nodes
	if _, ok := n.Latest.Lookup(report.HostNodeID); !ok {
		return MapEndpoint2Pseudo(n, local)
	}

	id := MakeHostID(report.ExtractHostID(n.Node))
	result := NewDerivedNode(id, n.WithParents(report.EmptySets))
	result.Children = result.Children.Add(n)
	result.Shape = Circle
	return RenderableNodes{id: result}
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
	// Propagate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a pod_id (maybe
	// slightly out of sync reports, or its not in a pod), just drop it
	podID, ok := n.Node.Latest.Lookup(kubernetes.PodID)
	if !ok {
		return RenderableNodes{}
	}
	id := MakePodID(podID)

	// Add container-<id> key to NMD, which will later be counted to produce the
	// minor label
	result := NewRenderableNodeWith(id, "", "", podID, n.WithParents(report.EmptySets))
	result.Counters = result.Counters.Add(ContainersKey, 1)

	// Due to a bug in kubernetes, addon pods on the master node are not returned
	// from the API. This is a workaround until
	// https://github.com/kubernetes/kubernetes/issues/14738 is fixed.
	if s := strings.SplitN(podID, "/", 2); len(s) == 2 {
		result.LabelMajor = s[1]
		result.Node = result.Node.WithLatests(map[string]string{
			kubernetes.Namespace: s[0],
			kubernetes.PodName:   s[1],
		})
	}

	result.Shape = Heptagon
	result.Children = result.Children.Add(n)
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
func MapPod2Service(pod RenderableNode, _ report.Networks) RenderableNodes {
	// Propagate all pseudo nodes
	if pod.Pseudo {
		return RenderableNodes{pod.ID: pod}
	}

	// Otherwise, if some some reason the pod doesn't have a service_ids (maybe
	// slightly out of sync reports, or its not in a service), just drop it
	ids, ok := pod.Node.Latest.Lookup(kubernetes.ServiceIDs)
	if !ok {
		return RenderableNodes{}
	}

	result := RenderableNodes{}
	for _, serviceID := range strings.Fields(ids) {
		id := MakeServiceID(serviceID)
		node := NewDerivedNode(id, pod.WithParents(report.EmptySets))
		node.Node.Counters = node.Node.Counters.Add(podsKey, 1)
		node.Children = node.Children.Add(pod)
		node.Shape = Heptagon
		node.Stack = true
		result[id] = node
	}
	return result
}

// MapContainer2Hostname maps container RenderableNodes to 'hostname' renderabled nodes..
func MapContainer2Hostname(n RenderableNode, _ report.Networks) RenderableNodes {
	// Propagate all pseudo nodes
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	// Otherwise, if some some reason the container doesn't have a hostname
	// (maybe slightly out of sync reports), just drop it
	id, ok := n.Node.Latest.Lookup(docker.ContainerHostname)
	if !ok {
		return RenderableNodes{}
	}

	result := NewDerivedNode(id, n)
	result.LabelMajor = id
	result.Rank = id

	// Add container id key to the counters, which will later be counted to produce the minor label
	result.Counters = result.Counters.Add(ContainersKey, 1)
	result.Node.Topology = "container_hostname"
	result.Node.ID = id
	result.Children = result.Children.Add(n)
	result.Shape = Hexagon
	result.Stack = true
	return RenderableNodes{id: result}
}

// MapCountContainers maps 1:1 container image nodes, counting
// the number of containers grouped together and putting
// that info in the minor label.
func MapCountContainers(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	output := n.Copy()
	containers, _ := n.Node.Counters.Lookup(ContainersKey)
	if containers == 1 {
		output.LabelMinor = "1 container"
	} else {
		output.LabelMinor = fmt.Sprintf("%d containers", containers)
	}
	return RenderableNodes{output.ID: output}
}

// MapCountPods maps 1:1 service nodes, counting the number of pods grouped
// together and putting that info in the minor label.
func MapCountPods(n RenderableNode, _ report.Networks) RenderableNodes {
	if n.Pseudo {
		return RenderableNodes{n.ID: n}
	}

	output := n.Copy()
	pods, _ := n.Node.Counters.Lookup(podsKey)
	if pods == 1 {
		output.LabelMinor = "1 pod"
	} else {
		output.LabelMinor = fmt.Sprintf("%d pods", pods)
	}
	return RenderableNodes{output.ID: output}
}
