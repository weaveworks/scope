package expected

import (
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

// Exported for testing.
var (
	circle   = "circle"
	square   = "square"
	heptagon = "heptagon"
	hexagon  = "hexagon"
	cloud    = "cloud"

	// Helper to make a report.node with some common options
	node = func(topology string) func(id string, adjacent ...string) report.Node {
		return func(id string, adjacent ...string) report.Node {
			n := report.MakeNode().WithID(id).WithTopology(topology)
			for _, a := range adjacent {
				n = n.WithAdjacent(a)
			}
			return n
		}
	}
	pseudo         = node(render.Pseudo)
	endpoint       = node("") // TODO: endpoints don't have a topology for some reason? Not being tagged?
	process        = node(report.Process)
	container      = node(report.Container)
	containerImage = node(report.ContainerImage)
	pod            = node(report.Pod)
	service        = node(report.Service)
	host           = node(report.Host)

	RenderedEndpoints = report.Nodes{
		fixture.Client54001NodeID:    endpoint(fixture.Client54001NodeID, fixture.Server80NodeID),
		fixture.Client54002NodeID:    endpoint(fixture.Client54002NodeID, fixture.Server80NodeID),
		fixture.Server80NodeID:       endpoint(fixture.Server80NodeID),
		fixture.UnknownClient1NodeID: endpoint(fixture.UnknownClient1NodeID, fixture.Server80NodeID),
		fixture.UnknownClient2NodeID: endpoint(fixture.UnknownClient2NodeID, fixture.Server80NodeID),
		fixture.UnknownClient3NodeID: endpoint(fixture.UnknownClient3NodeID, fixture.Server80NodeID),
		fixture.RandomClientNodeID:   endpoint(fixture.RandomClientNodeID, fixture.Server80NodeID),
		fixture.NonContainerNodeID:   endpoint(fixture.NonContainerNodeID, fixture.GoogleEndpointNodeID),
		fixture.GoogleEndpointNodeID: endpoint(fixture.GoogleEndpointNodeID),
	}.Prune()

	RenderedProcesses = report.Nodes{
		fixture.ClientProcess1NodeID:      process(fixture.ClientProcess1NodeID, fixture.ServerProcessNodeID),
		fixture.ClientProcess2NodeID:      process(fixture.ClientProcess2NodeID, fixture.ServerProcessNodeID),
		fixture.ServerProcessNodeID:       process(fixture.ServerProcessNodeID),
		fixture.NonContainerProcessNodeID: process(fixture.NonContainerProcessNodeID, render.OutgoingInternetID),
		unknownPseudoNode1ID:              pseudo(unknownPseudoNode1ID, fixture.ServerProcessNodeID),
		unknownPseudoNode2ID:              pseudo(unknownPseudoNode2ID, fixture.ServerProcessNodeID),
		render.IncomingInternetID:         pseudo(render.IncomingInternetID, fixture.ServerProcessNodeID),
		render.OutgoingInternetID:         pseudo(render.OutgoingInternetID),
	}.Prune()

	unknownPseudoNode1ID = render.MakePseudoNodeID(fixture.UnknownClient1IP)
	unknownPseudoNode2ID = render.MakePseudoNodeID(fixture.UnknownClient3IP)

	RenderedProcessNames = report.Nodes{
		fixture.Client1Name:       process(fixture.Client1Name, fixture.ServerName),
		fixture.ServerName:        process(fixture.ServerName),
		fixture.NonContainerName:  process(fixture.NonContainerName, render.OutgoingInternetID),
		unknownPseudoNode1ID:      pseudo(unknownPseudoNode1ID, fixture.ServerName),
		unknownPseudoNode2ID:      pseudo(unknownPseudoNode2ID, fixture.ServerName),
		render.IncomingInternetID: pseudo(render.IncomingInternetID, fixture.ServerName),
		render.OutgoingInternetID: pseudo(render.OutgoingInternetID),
	}.Prune()

	uncontainedServerID = render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostID)

	RenderedContainers = report.Nodes{
		fixture.ClientContainerNodeID: container(fixture.ClientContainerNodeID, fixture.ServerContainerNodeID),
		fixture.ServerContainerNodeID: container(fixture.ServerContainerNodeID),
		uncontainedServerID:           pseudo(uncontainedServerID, render.OutgoingInternetID),
		// unknownPseudoNode1ID:      pseudo(unknownPseudoNode1ID, fixture.ServerContainerNodeID),
		// unknownPseudoNode2ID:      pseudo(unknownPseudoNode2ID, fixture.ServerContainerNodeID),
		render.IncomingInternetID: pseudo(render.IncomingInternetID, fixture.ServerContainerNodeID),
		render.OutgoingInternetID: pseudo(render.OutgoingInternetID),
	}.Prune()

	RenderedContainerImages = report.Nodes{
		fixture.ClientContainerImageNodeID: containerImage(fixture.ClientContainerImageNodeID, fixture.ServerContainerImageNodeID),
		fixture.ServerContainerImageNodeID: containerImage(fixture.ServerContainerImageNodeID),
		uncontainedServerID:                pseudo(uncontainedServerID, render.OutgoingInternetID),
		// unknownPseudoNode1ID:      pseudo(unknownPseudoNode1ID, fixture.ServerContainerImageNodeID),
		// unknownPseudoNode2ID:      pseudo(unknownPseudoNode2ID, fixture.ServerContainerImageNodeID),
		render.IncomingInternetID: pseudo(render.IncomingInternetID, fixture.ServerContainerImageNodeID),
		render.OutgoingInternetID: pseudo(render.OutgoingInternetID),
	}.Prune()

	RenderedPods = report.Nodes{
		fixture.ClientPodNodeID: pod(fixture.ClientPodNodeID, fixture.ServerPodNodeID),
		fixture.ServerPodNodeID: pod(fixture.ServerPodNodeID),
		uncontainedServerID:     pseudo(uncontainedServerID, render.OutgoingInternetID),
		// unknownPseudoNode1ID:      pseudo(unknownPseudoNode1ID, fixture.ServerPodNodeID),
		// unknownPseudoNode2ID:      pseudo(unknownPseudoNode2ID, fixture.ServerPodNodeID),
		render.IncomingInternetID: pseudo(render.IncomingInternetID, fixture.ServerPodNodeID),
		render.OutgoingInternetID: pseudo(render.OutgoingInternetID),
	}.Prune()

	RenderedHosts = report.Nodes{
		fixture.ClientHostNodeID:  host(fixture.ClientHostNodeID, fixture.ServerHostNodeID),
		fixture.ServerHostNodeID:  host(fixture.ServerHostNodeID, render.OutgoingInternetID),
		unknownPseudoNode1ID:      pseudo(unknownPseudoNode1ID, fixture.ServerHostNodeID),
		unknownPseudoNode2ID:      pseudo(unknownPseudoNode2ID, fixture.ServerHostNodeID),
		render.IncomingInternetID: pseudo(render.IncomingInternetID, fixture.ServerHostNodeID),
		render.OutgoingInternetID: pseudo(render.OutgoingInternetID),
	}.Prune()

	RenderedPodServices = report.Nodes{
		fixture.ServiceNodeID: service(fixture.ServiceNodeID, fixture.ServiceNodeID),
		uncontainedServerID:   pseudo(uncontainedServerID, render.OutgoingInternetID),
		// unknownPseudoNode1ID:      pseudo(unknownPseudoNode1ID, fixture.ServiceID),
		// unknownPseudoNode2ID:      pseudo(unknownPseudoNode2ID, fixture.ServiceID),
		render.IncomingInternetID: pseudo(render.IncomingInternetID, fixture.ServiceNodeID),
		render.OutgoingInternetID: pseudo(render.OutgoingInternetID),
	}.Prune()
)

func newu64(value uint64) *uint64 { return &value }
