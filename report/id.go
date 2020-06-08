package report

import (
	"net"
	"strconv"
	"strings"
)

// Delimiters are used to separate parts of node IDs, to guarantee uniqueness
// in particular contexts.
const (
	// ScopeDelim is a general-purpose delimiter used within node IDs to
	// separate different contextual scopes. Different topologies have
	// different key structures.
	ScopeDelim = ";"

	// EdgeDelim separates two node IDs when they need to exist in the same key.
	// Concretely, it separates node IDs in keys that represent edges.
	EdgeDelim = "|"

	// Key added to nodes to prevent them being joined with conntracked connections
	DoesNotMakeConnections = "does_not_make_connections"

	// WeaveOverlayPeerPrefix is the prefix for weave peers in the overlay network
	WeaveOverlayPeerPrefix = ""

	// DockerOverlayPeerPrefix is the prefix for docker peers in the overlay network
	DockerOverlayPeerPrefix = "docker_peer_"
)

// MakeEndpointNodeID produces an endpoint node ID from its composite parts.
func MakeEndpointNodeID(hostID, namespaceID, address, port string) string {
	addressIP := net.ParseIP(address)
	return makeAddressID(hostID, namespaceID, address, addressIP) + ScopeDelim + port
}

// MakeEndpointNodeIDB produces an endpoint node ID from its composite parts in binary, not strings.
func MakeEndpointNodeIDB(hostID string, namespaceID uint32, addressIP net.IP, port uint16) string {
	namespace := ""
	if namespaceID > 0 {
		namespace = strconv.FormatUint(uint64(namespaceID), 10)
	}
	return makeAddressID(hostID, namespace, addressIP.String(), addressIP) + ScopeDelim + strconv.Itoa(int(port))
}

// MakeAddressNodeID produces an address node ID from its composite parts.
func MakeAddressNodeID(hostID, address string) string {
	addressIP := net.ParseIP(address)
	return makeAddressID(hostID, "", address, addressIP)
}

// MakeAddressNodeIDB produces an address node ID from its composite parts, in binary not string.
func MakeAddressNodeIDB(hostID string, addressIP net.IP) string {
	return makeAddressID(hostID, "", addressIP.String(), addressIP)
}

func makeAddressID(hostID, namespaceID, address string, addressIP net.IP) string {
	var scope string

	// Loopback addresses and addresses explicitly marked as local get
	// scoped by hostID
	// Loopback addresses are also scoped by the networking
	// namespace if available, since they can clash.
	if addressIP != nil && LocalNetworks.Contains(addressIP) {
		scope = hostID
	} else if addressIP != nil && addressIP.IsLoopback() {
		scope = hostID
		if namespaceID != "" {
			scope += "-" + namespaceID
		}
	}

	return scope + ScopeDelim + address
}

// MakeScopedEndpointNodeID is like MakeEndpointNodeID, but it always
// prefixes the ID with a scope.
func MakeScopedEndpointNodeID(scope, address, port string) string {
	return scope + ScopeDelim + address + ScopeDelim + port
}

// MakeScopedAddressNodeID is like MakeAddressNodeID, but it always
// prefixes the ID witha scope.
func MakeScopedAddressNodeID(scope, address string) string {
	return scope + ScopeDelim + address
}

// MakeProcessNodeID produces a process node ID from its composite parts.
func MakeProcessNodeID(hostID, pid string) string {
	return hostID + ScopeDelim + pid
}

// MakeECSServiceNodeID produces an ECS Service node ID from its composite parts.
func MakeECSServiceNodeID(cluster, serviceName string) string {
	return cluster + ScopeDelim + serviceName
}

var (
	// MakeHostNodeID produces a host node ID from its composite parts.
	MakeHostNodeID = makeSingleComponentID("host")

	// ParseHostNodeID parses a host node ID
	ParseHostNodeID = parseSingleComponentID("host")

	// MakeContainerNodeID produces a container node ID from its composite parts.
	MakeContainerNodeID = makeSingleComponentID("container")

	// ParseContainerNodeID parses a container node ID
	ParseContainerNodeID = parseSingleComponentID("container")

	// MakeContainerImageNodeID produces a container image node ID from its composite parts.
	MakeContainerImageNodeID = makeSingleComponentID("container_image")

	// ParseContainerImageNodeID parses a container image node ID
	ParseContainerImageNodeID = parseSingleComponentID("container_image")

	// MakePodNodeID produces a pod node ID from its composite parts.
	MakePodNodeID = makeSingleComponentID("pod")

	// ParsePodNodeID parses a pod node ID
	ParsePodNodeID = parseSingleComponentID("pod")

	// MakeServiceNodeID produces a service node ID from its composite parts.
	MakeServiceNodeID = makeSingleComponentID("service")

	// ParseServiceNodeID parses a service node ID
	ParseServiceNodeID = parseSingleComponentID("service")

	// MakeDeploymentNodeID produces a deployment node ID from its composite parts.
	MakeDeploymentNodeID = makeSingleComponentID("deployment")

	// ParseDeploymentNodeID parses a deployment node ID
	ParseDeploymentNodeID = parseSingleComponentID("deployment")

	// MakeReplicaSetNodeID produces a replica set node ID from its composite parts.
	MakeReplicaSetNodeID = makeSingleComponentID("replica_set")

	// ParseReplicaSetNodeID parses a replica set node ID
	ParseReplicaSetNodeID = parseSingleComponentID("replica_set")

	// MakeDaemonSetNodeID produces a replica set node ID from its composite parts.
	MakeDaemonSetNodeID = makeSingleComponentID("daemonset")

	// ParseDaemonSetNodeID parses a daemon set node ID
	ParseDaemonSetNodeID = parseSingleComponentID("daemonset")

	// MakeStatefulSetNodeID produces a statefulset node ID from its composite parts.
	MakeStatefulSetNodeID = makeSingleComponentID("statefulset")

	// ParseStatefulSetNodeID parses a statefulset node ID
	ParseStatefulSetNodeID = parseSingleComponentID("statefulset")

	// MakeCronJobNodeID produces a cronjob node ID from its composite parts.
	MakeCronJobNodeID = makeSingleComponentID("cronjob")

	// ParseCronJobNodeID parses a cronjob node ID
	ParseCronJobNodeID = parseSingleComponentID("cronjob")

	// MakeJobNodeID produces a job node ID from its composite parts.
	MakeJobNodeID = makeSingleComponentID("job")

	// ParseJobNodeID parses a job node ID
	ParseJobNodeID = parseSingleComponentID("job")

	// MakeNamespaceNodeID produces a namespace node ID from its composite parts.
	MakeNamespaceNodeID = makeSingleComponentID("namespace")

	// ParseNamespaceNodeID parses a namespace set node ID
	ParseNamespaceNodeID = parseSingleComponentID("namespace")

	// MakeECSTaskNodeID produces a ECSTask node ID from its composite parts.
	MakeECSTaskNodeID = makeSingleComponentID("ecs_task")

	// ParseECSTaskNodeID parses a ECSTask node ID
	ParseECSTaskNodeID = parseSingleComponentID("ecs_task")

	// MakeSwarmServiceNodeID produces a Swarm service node ID from its composite parts.
	MakeSwarmServiceNodeID = makeSingleComponentID("swarm_service")

	// ParseSwarmServiceNodeID parses a Swarm service node ID
	ParseSwarmServiceNodeID = parseSingleComponentID("swarm_service")

	// MakePersistentVolumeNodeID produces a Persistent Volume node ID from its composite parts.
	MakePersistentVolumeNodeID = makeSingleComponentID("persistent_volume")

	// ParsePersistentVolumeNodeID parses a Persistent Volume node ID
	ParsePersistentVolumeNodeID = parseSingleComponentID("persistent_volume")

	// MakePersistentVolumeClaimNodeID produces a Persistent Volume Claim node ID from its composite parts.
	MakePersistentVolumeClaimNodeID = makeSingleComponentID("persistent_volume_claim")

	// ParsePersistentVolumeClaimNodeID parses a Persistent Volume Claim node ID
	ParsePersistentVolumeClaimNodeID = parseSingleComponentID("persistent_volume_claim")

	// MakeStorageClassNodeID produces a storage class node ID from its composite parts.
	MakeStorageClassNodeID = makeSingleComponentID("storage_class")

	// ParseStorageClassNodeID parses a storage class node ID
	ParseStorageClassNodeID = parseSingleComponentID("storage_class")

	// MakeVolumeSnapshotNodeID produces a volume snapshot node ID from its composite parts.
	MakeVolumeSnapshotNodeID = makeSingleComponentID("volume_snapshot")

	// ParseVolumeSnapshotNodeID parses a volume snapshot node ID
	ParseVolumeSnapshotNodeID = parseSingleComponentID("volume_snapshot")

	// MakeVolumeSnapshotDataNodeID produces a volume snapshot data node ID from its composite parts.
	MakeVolumeSnapshotDataNodeID = makeSingleComponentID("volume_snapshot_data")

	// ParseVolumeSnapshotDataNodeID parses a volume snapshot data node ID
	ParseVolumeSnapshotDataNodeID = parseSingleComponentID("volume_snapshot_data")
)

// makeSingleComponentID makes a single-component node id encoder
func makeSingleComponentID(tag string) func(string) string {
	return func(id string) string {
		return id + ScopeDelim + "<" + tag + ">"
	}
}

// parseSingleComponentID makes a single-component node id decoder
func parseSingleComponentID(tag string) func(string) (string, bool) {
	return func(id string) (string, bool) {
		field0, field1, ok := split2(id, ScopeDelim)
		if !ok || field1 != "<"+tag+">" {
			return "", false
		}
		return field0, true
	}
}

// MakeOverlayNodeID produces an overlay topology node ID from a router peer's
// prefix and name, which is assumed to be globally unique.
func MakeOverlayNodeID(peerPrefix, peerName string) string {
	return "#" + peerPrefix + peerName
}

// ParseOverlayNodeID produces the overlay type and peer name.
func ParseOverlayNodeID(id string) (overlayPrefix string, peerName string) {

	if !strings.HasPrefix(id, "#") {
		// Best we can do
		return "", ""
	}

	id = id[1:]

	if strings.HasPrefix(id, DockerOverlayPeerPrefix) {
		return DockerOverlayPeerPrefix, id[len(DockerOverlayPeerPrefix):]
	}

	return WeaveOverlayPeerPrefix, id
}

// Split a string s into two parts separated by sep.
func split2(s, sep string) (s1, s2 string, ok bool) {
	// Not using strings.SplitN() to avoid a heap allocation
	pos := strings.Index(s, sep)
	if pos == -1 {
		return "", "", false
	}
	return s[:pos], s[pos+1:], true
}

// NodeIDType returns the type of a node ID - e.g. process, pod, endpoint
func NodeIDType(nodeID string) (string, bool) {
	if _, tag, ok := ParseNodeID(nodeID); ok {
		switch {
		case len(tag) >= 2 && tag[0] == '<' && tag[len(tag)-1] == '>':
			return tag[1 : len(tag)-1], true
		case len(tag) >= 1 && tag[0] >= '0' && tag[0] <= '9':
			return Endpoint, true
		}
	}
	return "", false
}

// ParseNodeID produces the id and tag of a single-component node ID.
func ParseNodeID(nodeID string) (id string, tag string, ok bool) {
	return split2(nodeID, ScopeDelim)
}

// ParseEndpointNodeID produces the scope, address, and port and remainder.
// Note that scope may be blank.
func ParseEndpointNodeID(endpointNodeID string) (scope, address, port string, ok bool) {
	// Not using strings.SplitN() to avoid a heap allocation
	first := strings.Index(endpointNodeID, ScopeDelim)
	if first == -1 {
		return "", "", "", false
	}
	second := strings.Index(endpointNodeID[first+1:], ScopeDelim)
	if second == -1 {
		return "", "", "", false
	}
	return endpointNodeID[:first], endpointNodeID[first+1 : first+1+second], endpointNodeID[first+1+second+1:], true
}

// ParseAddressNodeID produces the host ID, address from an address node ID.
func ParseAddressNodeID(addressNodeID string) (hostID, address string, ok bool) {
	return split2(addressNodeID, ScopeDelim)
}

// ParseProcessNodeID produces the host ID and PID from a process node ID.
func ParseProcessNodeID(processNodeID string) (hostID, pid string, ok bool) {
	return split2(processNodeID, ScopeDelim)
}

// ParseECSServiceNodeID produces the cluster, service name from an ECS Service node ID
func ParseECSServiceNodeID(ecsServiceNodeID string) (cluster, serviceName string, ok bool) {
	cluster, serviceName, ok = split2(ecsServiceNodeID, ScopeDelim)
	if !ok {
		return "", "", false
	}
	// In previous versions, ECS Service node IDs were of form serviceName + "<ecs_service>".
	// For backwards compatibility, we should still return a sensical serviceName for these cases.
	if serviceName == "<ecs_service>" {
		return "unknown", cluster, true
	}
	return cluster, serviceName, true
}

// ExtractHostID extracts the host id from Node
func ExtractHostID(m Node) string {
	hostNodeID, _ := m.Latest.Lookup(HostNodeID)
	hostID, _ := ParseHostNodeID(hostNodeID)
	return hostID
}

// IsLoopback ascertains if an address comes from a loopback interface.
func IsLoopback(address string) bool {
	ip := net.ParseIP(address)
	return ip != nil && ip.IsLoopback()
}

// IsPauseImageName indicates whether an image name corresponds to a
// kubernetes pause container image.
func IsPauseImageName(imageName string) bool {
	return strings.Contains(imageName, "google_containers/pause") ||
		strings.Contains(imageName, "k8s.gcr.io/pause") ||
		strings.Contains(imageName, "eks/pause")
}
