package render

import (
	"context"
	"regexp"
	"strings"

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
		state, ok := n.Latest.Lookup(report.DockerContainerState)
		return !ok || state != report.StateDeleted
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

func (c connectionJoin) Render(ctx context.Context, rpt report.Report) Nodes {
	inputNodes := TopologySelector(c.topology).Render(ctx, rpt).Nodes
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
		}, c.topology).Render(ctx, rpt)
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

type containerWithImageNameRenderer struct{}

// Render produces a container graph where the the latest metadata contains the
// container image name, if found.
func (r containerWithImageNameRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	containers := ContainerRenderer.Render(ctx, rpt)
	images := SelectContainerImage.Render(ctx, rpt)

	outputs := make(report.Nodes, len(containers.Nodes))
	for id, c := range containers.Nodes {
		outputs[id] = c
		imageID, ok := c.Latest.Lookup(report.DockerImageID)
		if !ok {
			continue
		}
		image, ok := images.Nodes[report.MakeContainerImageNodeID(imageID)]
		if !ok {
			continue
		}
		imageNodeID := containerImageNodeID(image)
		if imageNodeID == "" {
			continue
		}

		c.Latest = c.Latest.Propagate(image.Latest, report.DockerImageName, report.DockerImageTag,
			report.DockerImageSize, report.DockerImageVirtualSize, report.DockerImageLabelPrefix+"works.weave.role")

		c.Parents = c.Parents.
			Delete(report.ContainerImage).
			AddString(report.ContainerImage, imageNodeID)
		outputs[id] = c
	}
	return Nodes{Nodes: outputs, Filtered: containers.Filtered}
}

// ContainerWithImageNameRenderer is a Renderer which produces a container
// graph where the ranks are the image names, not their IDs
var ContainerWithImageNameRenderer = Memoise(containerWithImageNameRenderer{})

// ContainerImageRenderer produces a graph where each node is a container image
// with the original containers as children
var ContainerImageRenderer = Memoise(FilterEmpty(report.Container,
	containerImageRenderer{},
))

// ContainerHostnameRenderer is a Renderer which produces a renderable container
// by hostname graph..
//
// not memoised
var ContainerHostnameRenderer = FilterEmpty(report.Container,
	containerHostnameRenderer{},
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
	_, isInHostNetwork := m.Latest.Lookup(report.DockerIsInHostNetwork)
	if doesntMakeConnections || isInHostNetwork {
		return nil
	}

	result := []string{}
	if addrs, ok := m.Sets.Lookup(report.DockerContainerIPsWithScopes); ok {
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
	ports, _ := m.Sets.Lookup(report.DockerContainerPorts)
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
	if containerID, ok := n.Latest.Lookup(report.DockerContainerID); ok {
		id = report.MakeContainerNodeID(containerID)
		node = NewDerivedNode(id, n).WithTopology(report.Container)
	} else {
		hostID, _, _ := report.ParseProcessNodeID(n.ID)
		id = MakePseudoNodeID(UncontainedID, hostID)
		node = NewDerivedPseudoNode(id, n)
	}
	return node
}

// containerImageRenderer produces a graph where each node is a container image
// with the original containers as children
type containerImageRenderer struct{}

func (m containerImageRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	containers := ContainerWithImageNameRenderer.Render(ctx, rpt)
	images := rpt.ContainerImage.Nodes
	ret := newJoinResults(nil)

	for _, n := range containers.Nodes {
		if n.Topology == Pseudo {
			ret.passThrough(n)
			continue
		}
		// If some some reason the container doesn't have a image_id, just drop it
		imageID, ok := n.Latest.Lookup(report.DockerImageID)
		if !ok {
			continue
		}
		id := containerImageNodeID(n)
		if id == "" {
			continue
		}
		ret.addWithCreate(n, id, func() report.Node {
			imageID = report.MakeContainerImageNodeID(imageID)
			imageNode, ok := images[imageID]
			if !ok {
				imageNode = report.MakeNode(imageID).WithTopology(report.ContainerImage)
			}
			return imageNode.WithID(id)
		})
	}
	return ret.result(containers)
}

func containerImageNodeID(n report.Node) string {
	imageName, ok := n.Latest.Lookup(report.DockerImageName)
	if !ok {
		return ""
	}

	parts := strings.SplitN(imageName, "/", 3)
	if len(parts) == 3 {
		imageName = strings.Join(parts[1:3], "/")
	}
	imageNameWithoutTag := strings.SplitN(imageName, ":", 2)[0]
	return report.MakeContainerImageNodeID(imageNameWithoutTag)
}

var containerHostnameTopology = MakeGroupNodeTopology(report.Container, report.DockerContainerHostname)

// containerHostnameRenderer collects containers by docker hostname
type containerHostnameRenderer struct{}

func (m containerHostnameRenderer) Render(ctx context.Context, rpt report.Report) Nodes {
	containers := ContainerWithImageNameRenderer.Render(ctx, rpt)
	ret := newJoinResults(nil)

	for _, n := range containers.Nodes {
		if n.Topology == Pseudo {
			ret.passThrough(n)
			continue
		}
		// If some some reason the container doesn't have a hostname, just drop it
		id, ok := n.Latest.Lookup(report.DockerContainerHostname)
		if !ok {
			continue
		}
		ret.addChildAndChildren(n, id, containerHostnameTopology)
	}
	return ret.result(containers)
}
