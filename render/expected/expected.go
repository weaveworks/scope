package expected

import (
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/fixture"
)

// Exported for testing.
var (
	circle         = "circle"
	square         = "square"
	heptagon       = "heptagon"
	hexagon        = "hexagon"
	cloud          = "cloud"
	cylinder       = "cylinder"
	dottedcylinder = "dottedcylinder"
	storagesheet   = "sheet"

	// Helper to make a report.node with some common options
	node = func(topology string) func(id string, adjacent ...string) report.Node {
		return func(id string, adjacent ...string) report.Node {
			n := report.MakeNode(id).WithTopology(topology)
			for _, a := range adjacent {
				n = n.WithAdjacent(a)
			}
			return n
		}
	}
	pseudo                = node(render.Pseudo)
	endpoint              = node(report.Endpoint)
	processNode           = node(report.Process)
	processNameNode       = node(render.MakeGroupNodeTopology(report.Process, process.Name))
	container             = node(report.Container)
	containerHostnameNode = node(render.MakeGroupNodeTopology(report.Container, docker.ContainerHostname))
	containerImage        = node(report.ContainerImage)
	pod                   = node(report.Pod)
	service               = node(report.Service)
	hostNode              = node(report.Host)
	persistentVolume      = node(report.PersistentVolume)
	persistentVolumeClaim = node(report.PersistentVolumeClaim)
	StorageClass          = node(report.StorageClass)
	volumeSnapshot        = node(report.VolumeSnapshot)
	volumeSnapshotData    = node(report.VolumeSnapshotData)

	UnknownPseudoNode1ID = render.MakePseudoNodeID(fixture.UnknownClient1IP)
	UnknownPseudoNode2ID = render.MakePseudoNodeID(fixture.UnknownClient3IP)

	unknownPseudoNode1 = func(adjacent ...string) report.Node {
		return pseudo(UnknownPseudoNode1ID, adjacent...).
			WithChildren(report.MakeIDList(
				fixture.UnknownClient1NodeID,
				fixture.UnknownClient2NodeID,
			))
	}
	unknownPseudoNode2 = func(adjacent ...string) report.Node {
		return pseudo(UnknownPseudoNode2ID, adjacent...).
			WithChildren(report.MakeIDList(
				fixture.UnknownClient3NodeID,
			))
	}

	theIncomingInternetNode = func(adjacent ...string) report.Node {
		return pseudo(render.IncomingInternetID, adjacent...).
			WithChildren(report.MakeIDList(
				fixture.RandomClientNodeID,
			))
	}

	theOutgoingInternetNode = pseudo(render.OutgoingInternetID).WithChildID(fixture.GoogleEndpointNodeID)

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
	}

	RenderedProcesses = report.Nodes{
		fixture.ClientProcess1NodeID: processNode(fixture.ClientProcess1NodeID, fixture.ServerProcessNodeID).
			WithLatests(map[string]string{
				report.HostNodeID: fixture.ClientHostNodeID,
				process.PID:       fixture.Client1PID,
				process.Name:      fixture.Client1Name,
			}).
			WithChildID(fixture.Client54001NodeID),

		fixture.ClientProcess2NodeID: processNode(fixture.ClientProcess2NodeID, fixture.ServerProcessNodeID).
			WithChildID(fixture.Client54002NodeID),

		fixture.ServerProcessNodeID: processNode(fixture.ServerProcessNodeID).
			WithChildID(fixture.Server80NodeID),

		fixture.NonContainerProcessNodeID: processNode(fixture.NonContainerProcessNodeID, render.OutgoingInternetID).
			WithChildID(fixture.NonContainerNodeID),

		// due to https://github.com/weaveworks/scope/issues/1323 we are dropping
		// all non-internet pseudo nodes for now.
		// UnknownPseudoNode1ID: unknownPseudoNode1(fixture.ServerProcessNodeID),
		// UnknownPseudoNode2ID: unknownPseudoNode2(fixture.ServerProcessNodeID),
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerProcessNodeID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	RenderedProcessNames = report.Nodes{
		fixture.Client1Name: processNameNode(fixture.Client1Name, fixture.ServerName).
			WithLatests(map[string]string{process.Name: fixture.Client1Name}).
			AddCounter(report.Process, 2).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
			)),

		fixture.ServerName: processNameNode(fixture.ServerName).
			WithLatests(map[string]string{process.Name: fixture.ServerName}).
			AddCounter(report.Process, 1).
			WithChildren(report.MakeIDList(
				fixture.Server80NodeID,
				fixture.ServerProcessNodeID,
			)),

		fixture.NonContainerName: processNameNode(fixture.NonContainerName, render.OutgoingInternetID).
			AddCounter(report.Process, 1).
			WithChildren(report.MakeIDList(
				fixture.NonContainerNodeID,
				fixture.NonContainerProcessNodeID,
			)),

		// due to https://github.com/weaveworks/scope/issues/1323 we are dropping
		// all non-internet pseudo nodes for now.
		// UnknownPseudoNode1ID:      unknownPseudoNode1(fixture.ServerName),
		// UnknownPseudoNode2ID:      unknownPseudoNode2(fixture.ServerName),
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerName),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	uncontainedServerID   = render.MakePseudoNodeID(render.UncontainedID, fixture.ServerHostID)
	uncontainedServerNode = pseudo(uncontainedServerID, render.OutgoingInternetID).WithChildren(report.MakeIDList(
		fixture.NonContainerNodeID,
		fixture.NonContainerProcessNodeID,
	))

	RenderedContainers = report.Nodes{
		fixture.ClientContainerNodeID: container(fixture.ClientContainerNodeID, fixture.ServerContainerNodeID).
			WithLatests(map[string]string{
				report.HostNodeID:    fixture.ClientHostNodeID,
				docker.ContainerID:   fixture.ClientContainerID,
				docker.ContainerName: fixture.ClientContainerName,
				docker.ImageName:     fixture.ClientContainerImageName,
			}).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
			)),

		fixture.ServerContainerNodeID: container(fixture.ServerContainerNodeID).
			WithChildren(report.MakeIDList(
				fixture.Server80NodeID,
				fixture.ServerProcessNodeID,
			)),

		fixture.ServerContainer2NodeID: container(fixture.ServerContainer2NodeID),

		uncontainedServerID:       uncontainedServerNode,
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerContainerNodeID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	RenderedContainerHostnames = report.Nodes{
		fixture.ClientContainerHostname: containerHostnameNode(fixture.ClientContainerHostname, fixture.ServerContainerHostname).
			WithLatests(map[string]string{
				docker.ContainerHostname: fixture.ClientContainerHostname,
			}).
			AddCounter(report.Container, 1).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
				fixture.ClientContainerNodeID,
			)),

		fixture.ServerContainerHostname: containerHostnameNode(fixture.ServerContainerHostname).
			WithLatests(map[string]string{
				docker.ContainerHostname: fixture.ServerContainerHostname,
			}).
			AddCounter(report.Container, 2).
			WithChildren(report.MakeIDList(
				fixture.Server80NodeID,
				fixture.ServerProcessNodeID,
				fixture.ServerContainerNodeID,
				fixture.ServerContainer2NodeID,
			)),

		uncontainedServerID:       uncontainedServerNode,
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerContainerHostname),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	ClientContainerImageNodeID = report.MakeContainerImageNodeID(fixture.ClientContainerImageName)
	ServerContainerImageNodeID = report.MakeContainerImageNodeID(fixture.ServerContainerImageName)

	RenderedContainerImages = report.Nodes{
		ClientContainerImageNodeID: containerImage(ClientContainerImageNodeID, ServerContainerImageNodeID).
			WithLatests(map[string]string{
				report.HostNodeID: fixture.ClientHostNodeID,
				docker.ImageID:    fixture.ClientContainerImageID,
				docker.ImageName:  fixture.ClientContainerImageName,
			}).
			AddCounter(report.Container, 1).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
				fixture.ClientContainerNodeID,
			)),

		ServerContainerImageNodeID: containerImage(ServerContainerImageNodeID).
			AddCounter(report.Container, 2).
			WithChildren(report.MakeIDList(
				fixture.Server80NodeID,
				fixture.ServerProcessNodeID,
				fixture.ServerContainerNodeID,
				fixture.ServerContainer2NodeID,
			)),

		uncontainedServerID:       uncontainedServerNode,
		render.IncomingInternetID: theIncomingInternetNode(ServerContainerImageNodeID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	UnmanagedServerID   = render.MakePseudoNodeID(render.UnmanagedID, fixture.ServerHostID)
	unmanagedServerNode = pseudo(UnmanagedServerID, render.OutgoingInternetID).WithChildren(report.MakeIDList(
		uncontainedServerID,
		fixture.NonContainerNodeID,
		fixture.NonContainerProcessNodeID,
		fixture.ServerContainer2NodeID,
	)).
		AddCounter(report.Container, 1).
		AddCounter(render.Pseudo, 1)

	RenderedPods = report.Nodes{
		fixture.ClientPodNodeID: pod(fixture.ClientPodNodeID, fixture.ServerPodNodeID).
			AddCounter(report.Container, 1).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
				fixture.ClientContainerNodeID,
			)),

		fixture.ServerPodNodeID: pod(fixture.ServerPodNodeID).
			AddCounter(report.Container, 1).
			WithChildren(report.MakeIDList(
				fixture.Server80NodeID,
				fixture.ServerProcessNodeID,
				fixture.ServerContainerNodeID,
			)),

		fixture.PersistentVolumeClaimNodeID: persistentVolumeClaim(fixture.PersistentVolumeClaimNodeID, fixture.PersistentVolumeNodeID).
			WithLatests(map[string]string{
				kubernetes.Name:             "pvc-6124",
				kubernetes.Namespace:        "ping",
				kubernetes.Status:           "bound",
				kubernetes.VolumeName:       "pongvolume",
				kubernetes.AccessModes:      "ReadWriteOnce",
				kubernetes.StorageClassName: "standard",
			}).WithChildID(fixture.PersistentVolumeNodeID),

		fixture.PersistentVolumeNodeID: persistentVolume(fixture.PersistentVolumeNodeID, fixture.VolumeSnapshotNodeID).
			WithLatests(map[string]string{
				kubernetes.Name:             "pongvolume",
				kubernetes.Namespace:        "ping",
				kubernetes.Status:           "bound",
				kubernetes.VolumeClaim:      "pvc-6124",
				kubernetes.AccessModes:      "ReadWriteOnce",
				kubernetes.StorageClassName: "standard",
				kubernetes.StorageDriver:    "iSCSI",
			}).WithChildID(fixture.VolumeSnapshotNodeID),

		fixture.StorageClassNodeID: StorageClass(fixture.StorageClassNodeID, fixture.PersistentVolumeClaimNodeID).
			WithLatests(map[string]string{
				kubernetes.Name:        "standard",
				kubernetes.Provisioner: "pong",
			}).WithChildID(fixture.PersistentVolumeClaimNodeID),

		fixture.VolumeSnapshotNodeID: volumeSnapshot(fixture.VolumeSnapshotNodeID, fixture.VolumeSnapshotDataNodeID).
			WithLatests(map[string]string{
				kubernetes.Name:         "vs-1234",
				kubernetes.Namespace:    "ping",
				kubernetes.VolumeClaim:  "pvc-6124",
				kubernetes.SnapshotData: "vsd-1234",
				kubernetes.VolumeName:   "pongvolume",
			}).WithChildID(fixture.VolumeSnapshotDataNodeID),

		fixture.VolumeSnapshotDataNodeID: volumeSnapshotData(fixture.VolumeSnapshotDataNodeID).
			WithLatests(map[string]string{
				kubernetes.Name:               "vsd-1234",
				kubernetes.VolumeName:         "pongvolume",
				kubernetes.VolumeSnapshotName: "vs-1234",
			}),

		UnmanagedServerID:         unmanagedServerNode,
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerPodNodeID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	RenderedPodServices = report.Nodes{
		fixture.ServiceNodeID: service(fixture.ServiceNodeID, fixture.ServiceNodeID).
			AddCounter(report.Pod, 2).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.Server80NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
				fixture.ServerProcessNodeID,
				fixture.ClientContainerNodeID,
				fixture.ServerContainerNodeID,
				fixture.ClientPodNodeID,
				fixture.ServerPodNodeID,
			)),

		UnmanagedServerID:         unmanagedServerNode,
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServiceNodeID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}

	RenderedHosts = report.Nodes{
		fixture.ClientHostNodeID: hostNode(fixture.ClientHostNodeID, fixture.ServerHostNodeID).
			WithLatests(map[string]string{
				report.HostName: fixture.ClientHostName,
			}).
			AddCounter(report.Container, 1).
			AddCounter(report.ContainerImage, 1).
			AddCounter(report.Pod, 1).
			AddCounter(report.Process, 2).
			WithChildren(report.MakeIDList(
				fixture.Client54001NodeID,
				fixture.Client54002NodeID,
				fixture.ClientProcess1NodeID,
				fixture.ClientProcess2NodeID,
				fixture.ClientContainerNodeID,
				ClientContainerImageNodeID,
				fixture.ClientPodNodeID,
			)),

		fixture.ServerHostNodeID: hostNode(fixture.ServerHostNodeID, render.OutgoingInternetID).
			AddCounter(report.Container, 2).
			AddCounter(report.ContainerImage, 1).
			AddCounter(report.Pod, 1).
			AddCounter(report.Process, 2).
			WithChildren(report.MakeIDList(
				fixture.Server80NodeID,
				fixture.NonContainerNodeID,
				fixture.ServerProcessNodeID,
				fixture.NonContainerProcessNodeID,
				fixture.ServerContainerNodeID,
				fixture.ServerContainer2NodeID,
				ServerContainerImageNodeID,
				fixture.ServerPodNodeID,
			)),

		// due to https://github.com/weaveworks/scope/issues/1323 we are dropping
		// all non-internet pseudo nodes for now.
		// UnknownPseudoNode1ID:      unknownPseudoNode1(fixture.ServerHostNodeID),
		// UnknownPseudoNode2ID:      unknownPseudoNode2(fixture.ServerHostNodeID),
		render.IncomingInternetID: theIncomingInternetNode(fixture.ServerHostNodeID),
		render.OutgoingInternetID: theOutgoingInternetNode,
	}
)

func newu64(value uint64) *uint64 { return &value }
