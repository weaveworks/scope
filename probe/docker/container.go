package docker

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	ContainerName          = "docker_container_name"
	ContainerCommand       = "docker_container_command"
	ContainerPorts         = "docker_container_ports"
	ContainerCreated       = "docker_container_created"
	ContainerNetworks      = "docker_container_networks"
	ContainerIPs           = "docker_container_ips"
	ContainerHostname      = "docker_container_hostname"
	ContainerIPsWithScopes = "docker_container_ips_with_scopes"
	ContainerState         = "docker_container_state"
	ContainerStateHuman    = "docker_container_state_human"
	ContainerUptime        = "docker_container_uptime"
	ContainerRestartCount  = "docker_container_restart_count"
	ContainerNetworkMode   = "docker_container_network_mode"

	NetworkRxDropped = "network_rx_dropped"
	NetworkRxBytes   = "network_rx_bytes"
	NetworkRxErrors  = "network_rx_errors"
	NetworkTxPackets = "network_tx_packets"
	NetworkTxDropped = "network_tx_dropped"
	NetworkRxPackets = "network_rx_packets"
	NetworkTxErrors  = "network_tx_errors"
	NetworkTxBytes   = "network_tx_bytes"

	MemoryMaxUsage = "docker_memory_max_usage"
	MemoryUsage    = "docker_memory_usage"
	MemoryFailcnt  = "docker_memory_failcnt"
	MemoryLimit    = "docker_memory_limit"

	CPUPercpuUsage       = "docker_cpu_per_cpu_usage"
	CPUUsageInUsermode   = "docker_cpu_usage_in_usermode"
	CPUTotalUsage        = "docker_cpu_total_usage"
	CPUUsageInKernelmode = "docker_cpu_usage_in_kernelmode"
	CPUSystemCPUUsage    = "docker_cpu_system_cpu_usage"

	NetworkModeHost = "host"

	LabelPrefix = "docker_label_"
	EnvPrefix   = "docker_env_"

	stopTimeout = 10
)

// These 'constants' are used for node states.
// We need to take pointers to them, so they are vars...
var (
	StateCreated    = "created"
	StateDead       = "dead"
	StateExited     = "exited"
	StatePaused     = "paused"
	StateRestarting = "restarting"
	StateRunning    = "running"
	StateDeleted    = "deleted"
)

// StatsGatherer gathers container stats
type StatsGatherer interface {
	Stats(docker.StatsOptions) error
}

// Container represents a Docker container
type Container interface {
	UpdateState(*docker.Container)

	ID() string
	Image() string
	PID() int
	Hostname() string
	GetNode() report.Node
	State() string
	StateString() string
	HasTTY() bool
	Container() *docker.Container
	StartGatheringStats(StatsGatherer) error
	StopGatheringStats()
	NetworkMode() (string, bool)
	NetworkInfo([]net.IP) report.Sets
}

type container struct {
	sync.RWMutex
	container              *docker.Container
	stopStats              chan<- bool
	latestStats            docker.Stats
	pendingStats           [60]docker.Stats
	numPending             int
	hostID                 string
	baseNode               report.Node
	noCommandLineArguments bool
	noEnvironmentVariables bool
}

// NewContainer creates a new Container
func NewContainer(c *docker.Container, hostID string, noCommandLineArguments bool, noEnvironmentVariables bool) Container {
	result := &container{
		container:              c,
		hostID:                 hostID,
		noCommandLineArguments: noCommandLineArguments,
		noEnvironmentVariables: noEnvironmentVariables,
	}
	result.baseNode = result.getBaseNode()
	return result
}

func (c *container) UpdateState(container *docker.Container) {
	c.Lock()
	defer c.Unlock()
	c.container = container
}

func (c *container) ID() string {
	return c.container.ID
}

func (c *container) Image() string {
	return trimImageID(c.container.Image)
}

func (c *container) PID() int {
	return c.container.State.Pid
}

func (c *container) Hostname() string {
	if c.container.Config.Domainname == "" {
		return c.container.Config.Hostname
	}

	return fmt.Sprintf("%s.%s", c.container.Config.Hostname,
		c.container.Config.Domainname)
}

func (c *container) HasTTY() bool {
	return c.container.Config.Tty
}

func (c *container) State() string {
	return c.container.State.String()
}

func (c *container) StateString() string {
	return c.container.State.StateString()
}

func (c *container) Container() *docker.Container {
	return c.container
}

func (c *container) StartGatheringStats(client StatsGatherer) error {
	c.Lock()
	defer c.Unlock()

	if c.stopStats != nil {
		return nil
	}
	done := make(chan bool)
	c.stopStats = done

	stats := make(chan *docker.Stats)
	opts := docker.StatsOptions{
		ID:     c.container.ID,
		Stats:  stats,
		Stream: true,
		Done:   done,
	}

	log.Debugf("docker container: collecting stats for %s", c.container.ID)

	go func() {
		if err := client.Stats(opts); err != nil && err != io.EOF && err != io.ErrClosedPipe {
			log.Errorf("docker container: error collecting stats for %s: %v", c.container.ID, err)
		}
	}()

	go func() {
		for s := range stats {
			c.Lock()
			if c.numPending >= len(c.pendingStats) {
				log.Warnf("docker container: dropping stats for %s", c.container.ID)
			} else {
				c.latestStats = *s
				c.pendingStats[c.numPending] = *s
				c.numPending++
			}
			c.Unlock()
		}
		log.Debugf("docker container: stopped collecting stats for %s", c.container.ID)
		c.Lock()
		if c.stopStats == done {
			c.stopStats = nil
		}
		c.Unlock()
	}()

	return nil
}

func (c *container) StopGatheringStats() {
	c.Lock()
	defer c.Unlock()
	if c.stopStats != nil {
		close(c.stopStats)
		c.stopStats = nil
	}
}

func (c *container) ports(localAddrs []net.IP) report.StringSet {
	if c.container.NetworkSettings == nil {
		return report.MakeStringSet()
	}

	ports := []string{}
	for port, bindings := range c.container.NetworkSettings.Ports {
		if len(bindings) == 0 {
			ports = append(ports, fmt.Sprintf("%s", port))
			continue
		}
		for _, b := range bindings {
			if b.HostIP != "0.0.0.0" {
				ports = append(ports, fmt.Sprintf("%s:%s->%s", b.HostIP, b.HostPort, port))
				continue
			}

			for _, ip := range localAddrs {
				if ip.To4() != nil {
					ports = append(ports, fmt.Sprintf("%s:%s->%s", ip, b.HostPort, port))
				}
			}
		}
	}

	return report.MakeStringSet(ports...)
}

func (c *container) NetworkMode() (string, bool) {
	c.RLock()
	defer c.RUnlock()
	if c.container.HostConfig != nil {
		return c.container.HostConfig.NetworkMode, true
	}
	return "", false
}

func addScopeToIPs(hostID string, ips []string) []string {
	ipsWithScopes := []string{}
	for _, ip := range ips {
		ipsWithScopes = append(ipsWithScopes, report.MakeAddressNodeID(hostID, ip))
	}
	return ipsWithScopes
}

func isIPv4(addr string) bool {
	ip := net.ParseIP(addr)
	return ip != nil && ip.To4() != nil
}

func (c *container) NetworkInfo(localAddrs []net.IP) report.Sets {
	c.RLock()
	defer c.RUnlock()

	ips := c.container.NetworkSettings.SecondaryIPAddresses
	if c.container.NetworkSettings.IPAddress != "" {
		ips = append(ips, c.container.NetworkSettings.IPAddress)
	}

	// For now, for the proof-of-concept, we just add networks as a set of
	// names. For the next iteration, we will probably want to create a new
	// Network topology, populate the network nodes with all of the details
	// here, and provide foreign key links from nodes to networks.
	networks := make([]string, 0, len(c.container.NetworkSettings.Networks))
	for name, settings := range c.container.NetworkSettings.Networks {
		networks = append(networks, name)
		if settings.IPAddress != "" {
			ips = append(ips, settings.IPAddress)
		}
	}

	// Filter out IPv6 addresses; nothing works with IPv6 yet
	ipv4s := []string{}
	for _, ip := range ips {
		if isIPv4(ip) {
			ipv4s = append(ipv4s, ip)
		}
	}
	// Treat all Docker IPs as local scoped.
	ipsWithScopes := addScopeToIPs(c.hostID, ipv4s)

	return report.MakeSets().
		Add(ContainerNetworks, report.MakeStringSet(networks...)).
		Add(ContainerPorts, c.ports(localAddrs)).
		Add(ContainerIPs, report.MakeStringSet(ipv4s...)).
		Add(ContainerIPsWithScopes, report.MakeStringSet(ipsWithScopes...))
}

func (c *container) memoryUsageMetric(stats []docker.Stats) report.Metric {
	var max float64
	samples := make([]report.Sample, len(stats))
	for i, s := range stats {
		samples[i].Timestamp = s.Read
		samples[i].Value = float64(s.MemoryStats.Usage)
		if float64(s.MemoryStats.Limit) > max {
			max = float64(s.MemoryStats.Limit)
		}
	}
	return report.MakeMetric(samples).WithMax(max)
}

func (c *container) cpuPercentMetric(stats []docker.Stats) report.Metric {
	if len(stats) < 2 {
		return report.MakeMetric(nil)
	}

	samples := make([]report.Sample, len(stats)-1)
	previous := stats[0]
	for i, s := range stats[1:] {
		// Copies from docker/api/client/stats.go#L205
		cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage - previous.CPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(s.CPUStats.SystemCPUUsage - previous.CPUStats.SystemCPUUsage)
		cpuPercent := 0.0
		if systemDelta > 0.0 && cpuDelta > 0.0 {
			cpuPercent = (cpuDelta / systemDelta) * 100.0
		}
		samples[i].Timestamp = s.Read
		samples[i].Value = cpuPercent
		previous = s
	}
	return report.MakeMetric(samples).WithMax(100.0)
}

func (c *container) metrics() report.Metrics {
	if c.numPending == 0 {
		return report.Metrics{}
	}
	pendingStats := c.pendingStats[:c.numPending]
	result := report.Metrics{
		MemoryUsage:   c.memoryUsageMetric(pendingStats),
		CPUTotalUsage: c.cpuPercentMetric(pendingStats),
	}

	// leave one stat to help with relative metrics
	c.pendingStats[0] = c.pendingStats[c.numPending-1]
	c.numPending = 1
	return result
}

func (c *container) env() map[string]string {
	result := map[string]string{}
	for _, value := range c.container.Config.Env {
		v := strings.SplitN(value, "=", 2)
		if len(v) != 2 {
			continue
		}
		result[v[0]] = v[1]
	}
	return result
}

func (c *container) getSanitizedCommand() string {
	result := c.container.Path
	if !c.noCommandLineArguments {
		result = result + " " + strings.Join(c.container.Args, " ")
	}
	return result
}

func (c *container) getBaseNode() report.Node {
	result := report.MakeNodeWith(report.MakeContainerNodeID(c.ID()), map[string]string{
		ContainerID:       c.ID(),
		ContainerCreated:  c.container.Created.Format(time.RFC3339Nano),
		ContainerCommand:  c.getSanitizedCommand(),
		ImageID:           c.Image(),
		ContainerHostname: c.Hostname(),
	}).WithParents(report.MakeSets().
		Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(c.Image()))),
	)
	result = result.AddPrefixPropertyList(LabelPrefix, c.container.Config.Labels)
	if !c.noEnvironmentVariables {
		result = result.AddPrefixPropertyList(EnvPrefix, c.env())
	}
	return result
}

func (c *container) controlsMap() map[string]report.NodeControlData {
	paused := c.container.State.Paused
	running := !paused && c.container.State.Running
	stopped := !paused && !running
	return map[string]report.NodeControlData{
		UnpauseContainer: {Dead: !paused},
		RestartContainer: {Dead: !running},
		StopContainer:    {Dead: !running},
		PauseContainer:   {Dead: !running},
		AttachContainer:  {Dead: !running},
		ExecContainer:    {Dead: !running},
		StartContainer:   {Dead: !stopped},
		RemoveContainer:  {Dead: !stopped},
	}
}

func (c *container) GetNode() report.Node {
	c.RLock()
	defer c.RUnlock()
	latest := map[string]string{
		ContainerName:       strings.TrimPrefix(c.container.Name, "/"),
		ContainerState:      c.StateString(),
		ContainerStateHuman: c.State(),
	}
	controls := c.controlsMap()

	if !c.container.State.Paused && c.container.State.Running {
		uptime := (mtime.Now().Sub(c.container.State.StartedAt) / time.Second) * time.Second
		networkMode := ""
		if c.container.HostConfig != nil {
			networkMode = c.container.HostConfig.NetworkMode
		}
		latest[ContainerUptime] = uptime.String()
		latest[ContainerRestartCount] = strconv.Itoa(c.container.RestartCount)
		latest[ContainerNetworkMode] = networkMode
	}

	result := c.baseNode.WithLatests(latest)
	result = result.WithLatestControls(controls)
	result = result.WithMetrics(c.metrics())
	return result
}

// ExtractContainerIPs returns the list of container IPs given a Node from the Container topology.
func ExtractContainerIPs(nmd report.Node) []string {
	v, _ := nmd.Sets.Lookup(ContainerIPs)
	return []string(v)
}

// ExtractContainerIPsWithScopes returns the list of container IPs, prepended
// with scopes, given a Node from the Container topology.
func ExtractContainerIPsWithScopes(nmd report.Node) []string {
	v, _ := nmd.Sets.Lookup(ContainerIPsWithScopes)
	return []string(v)
}

// ContainerIsStopped checks if the docker container is in one of our "stopped" states
func ContainerIsStopped(c Container) bool {
	state := c.StateString()
	return (state != StateRunning && state != StateRestarting && state != StatePaused)
}
