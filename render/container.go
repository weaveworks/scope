package render

import (
	"regexp"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
)

// Constants are used in the tests.
const (
	UncontainedID    = "uncontained"
	UncontainedMajor = "Uncontained"

	// Topology for IPs so we can differentiate them at the end
	IP = "IP"
)

// UncontainedIDPrefix is the prefix of uncontained pseudo nodes
var UncontainedIDPrefix = MakePseudoNodeID(UncontainedID, "")

// ContainerRenderer is a Renderer which produces a renderable container
// graph by merging the process graph and the container topology.
// NB We only want processes in container _or_ processes with network connections
// but we need to be careful to ensure we only include each edge once, by only
// including the ProcessRenderer once.
var ContainerRenderer = Memoise(MakeFilter(
	func(n report.Node) bool {
		// Drop deleted containers
		state, ok := n.Latest.Lookup(docker.ContainerState)
		return !ok || state != docker.StateDeleted
	},
	MakeReduce(
		MakeMap(
			MapProcess2Container,
			ProcessRenderer,
		),
		ConnectionJoin(MapContainer2IP, report.Container),
	),
))

const originalNodeID = "original_node_id"

// ConnectionJoin joins the given topology with connections from the
// endpoints topology, using the toIPs function to extract IPs from
// the nodes.
func ConnectionJoin(toIPs func(report.Node) []string, topology string) Renderer {
	return connectionJoin{toIPs: toIPs, topology: topology}
}

type connectionJoin struct {
	toIPs    func(report.Node) []string
	topology string
}

func (c connectionJoin) Render(rpt report.Report) Nodes {
	inputNodes := TopologySelector(c.topology).Render(rpt).Nodes
	// Collect all the IPs we are trying to map to, and which ID they map from
	var ipNodes = map[string]string{}
	for _, n := range inputNodes {
		for _, ip := range c.toIPs(n) {
			if _, exists := ipNodes[ip]; exists {
				// If an IP is shared between multiple nodes, we can't reliably
				// attribute an connection based on its IP
				ipNodes[ip] = "" // blank out the mapping so we don't use it
			} else {
				ipNodes[ip] = n.ID
			}
		}
	}
	return MapEndpoints(
		func(m report.Node) string {
			scope, addr, port, ok := report.ParseEndpointNodeID(m.ID)
			if !ok {
				return ""
			}
			id, found := ipNodes[report.MakeScopedEndpointNodeID(scope, addr, "")]
			// We also allow for joining on ip:port pairs.  This is
			// useful for connections to the host IPs which have been
			// port mapped to a container can only be unambiguously
			// identified with the port.
			if !found {
				id, found = ipNodes[report.MakeScopedEndpointNodeID(scope, addr, port)]
			}
			if !found || id == "" {
				return ""
			}
			// Not an IP we blanked out earlier.
			//
			// MapEndpoints is guaranteed to find a node with this id
			// (and hence not have to create one), since we got the id
			// from ipNodes, which is populated from c.topology, which
			// is where MapEndpoints will look.
			return id
		}, c.topology).Render(rpt)
}

// FilterEmpty is a Renderer which filters out nodes which have no children
// from the specified topology.
func FilterEmpty(topology string, r Renderer) Renderer {
	return MakeFilter(HasChildren(topology), r)
}

// HasChildren returns true if the node has no children from the specified
// topology.
func HasChildren(topology string) FilterFunc {
	return func(n report.Node) bool {
		count := 0
		n.Children.ForEach(func(child report.Node) {
			if child.Topology == topology {
				count++
			}
		})
		return count > 0
	}
}

type containerWithImageNameRenderer struct {
	Renderer
}

// Render produces a container graph where the the latest metadata contains the
// container image name, if found.
func (r containerWithImageNameRenderer) Render(rpt report.Report) Nodes {
	containers := r.Renderer.Render(rpt)
	images := SelectContainerImage.Render(rpt)

	outputs := make(report.Nodes, len(containers.Nodes))
	for id, c := range containers.Nodes {
		outputs[id] = c
		imageID, ok := c.Latest.Lookup(docker.ImageID)
		if !ok {
			continue
		}
		image, ok := images.Nodes[report.MakeContainerImageNodeID(imageID)]
		if !ok {
			continue
		}
		imageName, ok := image.Latest.Lookup(docker.ImageName)
		if !ok {
			continue
		}
		imageNameWithoutVersion := docker.ImageNameWithoutVersion(imageName)
		imageNodeID := report.MakeContainerImageNodeID(imageNameWithoutVersion)

		c = propagateLatest(docker.ImageName, image, c)
		c = propagateLatest(docker.ImageSize, image, c)
		c = propagateLatest(docker.ImageVirtualSize, image, c)
		c = propagateLatest(docker.ImageLabelPrefix+"works.weave.role", image, c)
		c.Parents = c.Parents.
			Delete(report.ContainerImage).
			Add(report.ContainerImage, report.MakeStringSet(imageNodeID))
		outputs[id] = c
	}
	return Nodes{Nodes: outputs, Filtered: containers.Filtered}
}

// ContainerWithImageNameRenderer is a Renderer which produces a container
// graph where the ranks are the image names, not their IDs
var ContainerWithImageNameRenderer = Memoise(containerWithImageNameRenderer{ContainerRenderer})

// ContainerImageRenderer is a Renderer which produces a renderable container
// image graph by merging the container graph and the container image topology.
var ContainerImageRenderer = Memoise(FilterEmpty(report.Container,
	MakeMap(
		MapContainerImage2Name,
		MakeReduce(
			MakeMap(
				MapContainer2ContainerImage,
				ContainerWithImageNameRenderer,
			),
			SelectContainerImage,
		),
	),
))

// ContainerHostnameRenderer is a Renderer which produces a renderable container
// by hostname graph..
//
// not memoised
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

// MapContainer2IP maps container nodes to their IP addresses (outputs
// multiple nodes).  This allows container to be joined directly with
// the endpoint topology.
func MapContainer2IP(m report.Node) []string {
	// if this container doesn't make connections, we can ignore it
	_, doesntMakeConnections := m.Latest.Lookup(report.DoesNotMakeConnections)
	// if this container belongs to the host's networking namespace
	// we cannot use its IP to attribute connections
	// (they could come from any other process on the host or DNAT-ed IPs)
	_, isInHostNetwork := m.Latest.Lookup(docker.IsInHostNetwork)
	if doesntMakeConnections || isInHostNetwork {
		return nil
	}

	result := []string{}
	if addrs, ok := m.Sets.Lookup(docker.ContainerIPsWithScopes); ok {
		for _, addr := range addrs {
			scope, addr, ok := report.ParseAddressNodeID(addr)
			if !ok {
				continue
			}
			// loopback addresses are shared among all namespaces
			// so we can't use them to attribute connections to a container
			if report.IsLoopback(addr) {
				continue
			}
			id := report.MakeScopedEndpointNodeID(scope, addr, "")
			result = append(result, id)
		}
	}

	// Also output all the host:port port mappings (see above comment).
	// In this case we assume this doesn't need a scope, as they are for host IPs.
	ports, _ := m.Sets.Lookup(docker.ContainerPorts)
	for _, portMapping := range ports {
		if mapping := portMappingMatch.FindStringSubmatch(portMapping); mapping != nil {
			ip, port := mapping[1], mapping[2]
			id := report.MakeScopedEndpointNodeID("", ip, port)
			result = append(result, id)
		}
	}

	return result
}

// MapProcess2Container maps process Nodes to container
// Nodes.
//
// Pseudo nodes are passed straight through.
//
// If this function is given a node without a docker_container_id, it
// will produce an "Uncontained" pseudo node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container, but without any Major or Minor labels.
// It does not have enough info to do that, and the resulting graph
// must be merged with a container graph to get that info.
func MapProcess2Container(n report.Node) report.Node {
	// Propagate pseudo nodes
	if n.Topology == Pseudo {
		return n
	}

	// Otherwise, if the process is not in a container, group it into
	// a per-host "Uncontained" node.
	var (
		id   string
		node report.Node
	)
	if containerID, ok := n.Latest.Lookup(docker.ContainerID); ok {
		id = report.MakeContainerNodeID(containerID)
		node = NewDerivedNode(id, n).WithTopology(report.Container)
	} else {
		hostID, _, _ := report.ParseProcessNodeID(n.ID)
		id = MakePseudoNodeID(UncontainedID, hostID)
		node = NewDerivedPseudoNode(id, n)
	}
	return node
}

// MapContainer2ContainerImage maps container Nodes to container
// image Nodes.
//
// Pseudo nodes are passed straight through.
//
// If this function is given a node without a docker_image_id
// it will drop that node.
//
// Otherwise, this function will produce a node with the correct ID
// format for a container image, but without any Major or Minor
// labels.  It does not have enough info to do that, and the resulting
// graph must be merged with a container image graph to get that info.
func MapContainer2ContainerImage(n report.Node) report.Node {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return n
	}

	// Otherwise, if some some reason the container doesn't have a image_id
	// (maybe slightly out of sync reports), just drop it
	imageID, ok := n.Latest.Lookup(docker.ImageID)
	if !ok {
		return report.Node{}
	}

	// Add container id key to the counters, which will later be
	// counted to produce the minor label
	id := report.MakeContainerImageNodeID(imageID)
	result := NewDerivedNode(id, n).WithTopology(report.ContainerImage)
	result.Counters = result.Counters.Add(n.Topology, 1)
	return result
}

// MapContainerImage2Name ignores image versions
func MapContainerImage2Name(n report.Node) report.Node {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return n
	}

	imageName, ok := n.Latest.Lookup(docker.ImageName)
	if !ok {
		return report.Node{}
	}

	imageNameWithoutVersion := docker.ImageNameWithoutVersion(imageName)
	n.ID = report.MakeContainerImageNodeID(imageNameWithoutVersion)

	return n
}

var containerHostnameTopology = MakeGroupNodeTopology(report.Container, docker.ContainerHostname)

// MapContainer2Hostname maps container Nodes to 'hostname' renderabled nodes..
func MapContainer2Hostname(n report.Node) report.Node {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return n
	}

	// Otherwise, if some some reason the container doesn't have a hostname
	// (maybe slightly out of sync reports), just drop it
	id, ok := n.Latest.Lookup(docker.ContainerHostname)
	if !ok {
		return report.Node{}
	}

	node := NewDerivedNode(id, n).WithTopology(containerHostnameTopology)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return node
}

// MapToEmpty removes all the attributes, children, etc, of a node. Useful when
// we just want to count the presence of nodes.
func MapToEmpty(n report.Node) report.Node {
	return report.MakeNode(n.ID).WithTopology(n.Topology)
}
