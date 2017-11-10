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
var UncontainedIDPrefix = MakePseudoNodeID(UncontainedID)

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
			ColorConnectedProcessRenderer,
		),
		ConnectionJoin(MapContainer2IP, SelectContainer),
	),
))

const originalNodeID = "original_node_id"

// ConnectionJoin joins the given renderer with connections from the
// endpoints topology, using the toIPs function to extract IPs from
// the nodes.
func ConnectionJoin(toIPs func(report.Node) []string, r Renderer) Renderer {
	return connectionJoin{toIPs: toIPs, r: r}
}

type connectionJoin struct {
	toIPs func(report.Node) []string
	r     Renderer
}

func (c connectionJoin) Render(rpt report.Report, dct Decorator) Nodes {
	local := LocalNetworks(rpt)
	inputNodes := c.r.Render(rpt, dct)
	endpoints := SelectEndpoint.Render(rpt, dct)

	// Collect all the IPs we are trying to map to, and which ID they map from
	var ipNodes = map[string]string{}
	for _, n := range inputNodes.Nodes {
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
	ret := newJoinResults()

	// Now look at all the endpoints and see which map to IP nodes
	for _, m := range endpoints.Nodes {
		scope, addr, port, ok := report.ParseEndpointNodeID(m.ID)
		if !ok {
			continue
		}
		// Nodes without a hostid may be pseudo nodes - if so, pass through to result
		if _, ok := m.Latest.Lookup(report.HostNodeID); !ok {
			if id, ok := externalNodeID(m, addr, local); ok {
				ret.addToResults(m, id, newPseudoNode)
				continue
			}
		}
		id, found := ipNodes[report.MakeScopedEndpointNodeID(scope, addr, "")]
		// We also allow for joining on ip:port pairs.  This is useful for
		// connections to the host IPs which have been port mapped to a
		// container can only be unambiguously identified with the port.
		if !found {
			id, found = ipNodes[report.MakeScopedEndpointNodeID(scope, addr, port)]
		}
		if found && id != "" { // not one we blanked out earlier
			ret.addToResults(m, id, func(id string) report.Node {
				return inputNodes.Nodes[id]
			})
		}
	}
	ret.copyUnmatched(inputNodes)
	ret.fixupAdjacencies(inputNodes)
	ret.fixupAdjacencies(endpoints)
	return ret.result()
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
func (r containerWithImageNameRenderer) Render(rpt report.Report, dct Decorator) Nodes {
	containers := r.Renderer.Render(rpt, dct)
	images := SelectContainerImage.Render(rpt, dct)

	outputs := report.Nodes{}
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
		node = propagateLatest(report.HostNodeID, n, node)
		node = propagateLatest(IsConnected, n, node)
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

// MapContainerImage2Name ignores image versions
func MapContainerImage2Name(n report.Node, _ report.Networks) report.Nodes {
	// Propagate all pseudo nodes
	if n.Topology == Pseudo {
		return report.Nodes{n.ID: n}
	}

	imageName, ok := n.Latest.Lookup(docker.ImageName)
	if !ok {
		return report.Nodes{}
	}

	imageNameWithoutVersion := docker.ImageNameWithoutVersion(imageName)
	n.ID = report.MakeContainerImageNodeID(imageNameWithoutVersion)

	if imageID, ok := report.ParseContainerImageNodeID(n.ID); ok {
		n.Sets = n.Sets.Add(docker.ImageID, report.MakeStringSet(imageID))
	}

	return report.Nodes{n.ID: n}
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
	node.Latest = node.Latest.Set(docker.ContainerHostname, timestamp, id)
	node.Counters = node.Counters.Add(n.Topology, 1)
	return report.Nodes{id: node}
}

// MapToEmpty removes all the attributes, children, etc, of a node. Useful when
// we just want to count the presence of nodes.
func MapToEmpty(n report.Node, _ report.Networks) report.Nodes {
	return report.Nodes{n.ID: report.MakeNode(n.ID).WithTopology(n.Topology)}
}
