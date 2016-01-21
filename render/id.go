package render

import (
	"strings"
)

// makeID is the generic ID maker
func makeID(prefix string, parts ...string) string {
	return strings.Join(append([]string{prefix}, parts...), ":")
}

// MakeEndpointID makes an endpoint node ID for rendered nodes.
func MakeEndpointID(hostID, addr, port string) string {
	return makeID("endpoint", hostID, addr, port)
}

// MakeProcessID makes a process node ID for rendered nodes.
func MakeProcessID(hostID, pid string) string {
	return makeID("process", hostID, pid)
}

// MakeAddressID makes an address node ID for rendered nodes.
func MakeAddressID(hostID, addr string) string {
	return makeID("address", hostID, addr)
}

// MakeContainerID makes a container node ID for rendered nodes.
func MakeContainerID(containerID string) string {
	return makeID("container", containerID)
}

// MakeContainerImageID makes a container image node ID for rendered nodes.
func MakeContainerImageID(imageID string) string {
	return makeID("container_image", imageID)
}

// MakePodID makes a pod node ID for rendered nodes.
func MakePodID(podID string) string {
	return makeID("pod", podID)
}

// MakeServiceID makes a service node ID for rendered nodes.
func MakeServiceID(serviceID string) string {
	return makeID("service", serviceID)
}

// MakeHostID makes a host node ID for rendered nodes.
func MakeHostID(hostID string) string {
	return makeID("host", hostID)
}

// MakePseudoNodeID produces a pseudo node ID from its composite parts,
// for use in rendered nodes.
func MakePseudoNodeID(parts ...string) string {
	return makeID("pseudo", parts...)
}

// MakeGroupID makes a node ID for grouped aggregate rendered nodes.
func MakeGroupID(key, val string) string {
	return makeID("group", key, val)
}
