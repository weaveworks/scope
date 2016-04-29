package render

import (
	"net"
	"regexp"
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/endpoint"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	UncontainedID    = "uncontained"
	UncontainedMajor = "Uncontained"

	// Topology for IPs so we can differentiate them at the end
	IP = "IP"
)

// ContainerRenderer is a Renderer which produces a renderable container
// graph by merging the process graph and the container topology.
// NB We only want processes in container _or_ processes with network connections
// but we need to be careful to ensure we only include each edge once, by only
// including the ProcessRenderer once.
var ContainerRenderer = MakeFilter(
	func(n report.Node) bool {
		// Drop deleted containers
		state, ok := n.Latest.Lookup(docker.ContainerState)
		return !ok || state != docker.StateDeleted
	},
	MakeReduce(
		MakeFilter(
			func(n report.Node) bool {
				// Drop unconnected pseudo nodes (could appear due to filtering)
				_, isConnected := n.Latest.Lookup(IsConnected)
				return n.Topology != Pseudo || isConnected
			},
			MakeMap(
				MapProcess2Container,
				ColorConnected(ProcessRenderer),
			),
		),

		// This mapper brings in short lived connections by joining with container IPs.
		// We need to be careful to ensure we only include each edge once.  Edges brought in
		// by the above renders will have a pid, so its enough to filter out any nodes with
		// pids.
		FilterUnconnected(MakeMap(
			MapIP2Container,
			MakeReduce(
				MakeMap(
					MapContainer2IP,
					SelectContainer,
				),
				MakeMap(
					MapEndpoint2IP,
					SelectEndpoint,
				),
			),
		)),

		SelectContainer,
	),
)

type containerWithImageNameRenderer struct {
	Renderer
}

// Render produces a process graph where the minor labels contain the
// container name, if found.  It also merges the image node metadata into the
// container metadata.
func (r containerWithImageNameRenderer) Render(rpt report.Report, dct Decorator) report.Nodes {
	containers := r.Renderer.Render(rpt, dct)
	images := SelectContainerImage.Render(rpt, dct)

	outputs := report.Nodes{}
	for id, c := range containers {
		outputs[id] = c
		imageID, ok := c.Latest.Lookup(docker.ImageID)
		if !ok {
			continue
		}
		image, ok := images[report.MakeContainerImageNodeID(imageID)]
		if !ok {
			continue
		}
		output := c.Copy()
		output.Latest = image.Latest.Merge(c.Latest)
		outputs[id] = output
	}
	return outputs
}

// ContainerWithImageNameRenderer is a Renderer which produces a container
// graph where the ranks are the image names, not their IDs
var ContainerWithImageNameRenderer = ApplyDecorators(containerWithImageNameRenderer{ContainerRenderer})

// ContainerImageRenderer is a Renderer which produces a renderable container
// image graph by merging the container graph and the container image topology.
var ContainerImageRenderer = FilterEmpty(report.Container,
	MakeReduce(
		MakeMap(
			MapContainer2ContainerImage,
			ContainerWithImageNameRenderer,
		),
		SelectContainerImage,
	),
)

// ContainerHostnameRenderer is a Renderer which produces a renderable container
// by hostname graph..
var ContainerHostnameRenderer = FilterEmpty(report.Container,
	MakeReduce(
		MakeMap(
			MapContainer2Hostname,
			ContainerWithImageNameRenderer,
		),
		// Grab *all* the hostnames, so we can count the number which were empty
		// for accurate stats.
		MakeMap(
			MapToEmpty,
			MakeMap(
				MapContainer2Hostname,
				ContainerRenderer,
			),
		),
	),
)

var portMappingMatch = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}):([0-9]+)->([0-9]+)/tcp`)

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

// MapContainer2IP maps container nodes to their IP addresses (outputs
// multiple nodes).  This allows container to be joined directly with
// the endpoint topology.
func MapContainer2IP(m report.Node, _ report.Networks) report.Nodes {
	containerID, ok := m.Latest.Lookup(docker.ContainerID)
	if !ok {
		return report.Nodes{}
	}

	// if this container doesn't make connections, we can ignore it
	_, doesntMakeConnections := m.Latest.Lookup(report.DoesNotMakeConnections)
	if doesntMakeConnections {
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
				WithCounters(map[string]int{IP: 1})

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
				WithCounters(map[string]int{IP: 1})

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
	if count, _ := n.Counters.Lookup(IP); count > 1 {
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
	// Propagate pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
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

	node := NewDerivedNode(id, n).WithTopology(MakeGroupNodeTopology(n.Topology, docker.ContainerHostname))
	node.Latest = node.Latest.
		Set(docker.ContainerHostname, timestamp, id).
		Delete(docker.ContainerName) // TODO(paulbellamy): total hack to render these by hostname instead.
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
}

// MapToEmpty removes all the attributes, children, etc, of a node. Useful when
// we just want to count the presence of nodes.
func MapToEmpty(n report.Node, _ report.Networks) report.Nodes {
	return report.Nodes{n.ID: report.MakeNode(n.ID).WithTopology(n.Topology)}
}
