package detailed

import (
	"strings"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/probe/host"
	"github.com/weaveworks/scope/probe/kubernetes"
	"github.com/weaveworks/scope/probe/overlay"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/render"
)

var labels = map[string]string{
	docker.CPUTotalUsage:         "CPU",
	docker.ContainerCommand:      "Command",
	docker.ContainerCreated:      "Created",
	docker.ContainerHostname:     "Hostname",
	docker.ContainerID:           "ID",
	docker.ContainerIPs:          "IPs",
	docker.ContainerPorts:        "Ports",
	docker.ContainerRestartCount: "Restart #",
	docker.ContainerState:        "State",
	docker.ContainerUptime:       "Uptime",
	docker.ImageID:               "Image ID",
	docker.MemoryUsage:           "Memory",
	host.CPUUsage:                "CPU",
	host.HostName:                "Hostname",
	host.KernelVersion:           "Kernel version",
	host.Load15:                  "Load (15m)",
	host.Load1:                   "Load (1m)",
	host.Load5:                   "Load (5m)",
	host.LocalNetworks:           "Local Networks",
	host.MemoryUsage:             "Memory",
	host.OS:                      "OS",
	host.Uptime:                  "Uptime",
	kubernetes.Namespace:         "Namespace",
	kubernetes.PodCreated:        "Created",
	kubernetes.PodID:             "ID",
	overlay.WeaveDNSHostname:     "Weave DNS Name",
	overlay.WeaveMACAddress:      "Weave MAC",
	process.CPUUsage:             "CPU",
	process.Cmdline:              "Command",
	process.MemoryUsage:          "Memory",
	process.PID:                  "PID",
	process.PPID:                 "Parent PID",
	process.Threads:              "# Threads",
	render.ContainersKey:         "# Containers",
}

// Label maps from the internal keys to the human-readable label for a piece
// of metadata/set/etc. If none is found the raw key will be returned.
func Label(key string) string {
	if label, ok := labels[key]; ok {
		return label
	}
	if strings.HasPrefix(key, "label_") {
		return strings.TrimPrefix(key, "label_")
	}
	return key
}
