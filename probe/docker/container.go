package docker

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/common/mtime"
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

// Exported for testing
var (
	DialStub          = net.Dial
	NewClientConnStub = newClientConn
)

func newClientConn(c net.Conn, r *bufio.Reader) ClientConn {
	return httputil.NewClientConn(c, r)
}

// ClientConn is exported for testing
type ClientConn interface {
	Do(req *http.Request) (resp *http.Response, err error)
	Close() error
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
	StartGatheringStats() error
	StopGatheringStats()
	NetworkMode() (string, bool)
	NetworkInfo([]net.IP) report.Sets
}

type container struct {
	sync.RWMutex
	container    *docker.Container
	statsConn    ClientConn
	latestStats  docker.Stats
	pendingStats [60]docker.Stats
	numPending   int
	hostID       string
	baseNode     report.Node
}

// NewContainer creates a new Container
func NewContainer(c *docker.Container, hostID string) Container {
	result := &container{
		container: c,
		hostID:    hostID,
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

func (c *container) StartGatheringStats() error {
	c.Lock()
	defer c.Unlock()

	if c.statsConn != nil {
		return fmt.Errorf("already gather stats for container %s", c.container.ID)
	}

	go func() {
		log.Infof("docker container: collecting stats for %s", c.container.ID)
		req, err := http.NewRequest("GET", fmt.Sprintf("/containers/%s/stats", c.container.ID), nil)
		if err != nil {
			log.Errorf("docker container: %v", err)
			return
		}
		req.Header.Set("User-Agent", "weavescope")

		url, err := url.Parse(endpoint)
		if err != nil {
			log.Errorf("docker container: %v", err)
			return
		}

		dial, err := DialStub(url.Scheme, url.Path)
		if err != nil {
			log.Errorf("docker container: %v", err)
			return
		}

		conn := NewClientConnStub(dial, nil)
		resp, err := conn.Do(req)
		if err != nil {
			log.Errorf("docker container: %v", err)
			return
		}
		defer resp.Body.Close()

		c.Lock()
		c.statsConn = conn
		c.Unlock()

		defer func() {
			c.Lock()
			defer c.Unlock()

			log.Infof("docker container: stopped collecting stats for %s", c.container.ID)
			c.statsConn = nil
		}()

		var stats docker.Stats
		// Use a buffer since the codec library doesn't implicitly do it
		bufReader := bufio.NewReader(resp.Body)
		decoder := codec.NewDecoder(bufReader, &codec.JsonHandle{})
		for err := decoder.Decode(&stats); err != io.EOF; err = decoder.Decode(&stats) {
			if err != nil {
				log.Errorf("docker container: error reading event, did container stop? %v", err)
				return
			}

			c.Lock()
			if c.numPending >= len(c.pendingStats) {
				log.Warnf("docker container: dropping stats.")
			} else {
				c.latestStats = stats
				c.pendingStats[c.numPending] = stats
				c.numPending++
			}
			c.Unlock()

			stats = docker.Stats{}
		}
	}()

	return nil
}

func (c *container) StopGatheringStats() {
	c.Lock()
	defer c.Unlock()

	if c.statsConn == nil {
		return
	}

	c.statsConn.Close()
	c.statsConn = nil
	return
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

	return report.EmptySets.
		Add(ContainerNetworks, report.MakeStringSet(networks...)).
		Add(ContainerPorts, c.ports(localAddrs)).
		Add(ContainerIPs, report.MakeStringSet(ipv4s...)).
		Add(ContainerIPsWithScopes, report.MakeStringSet(ipsWithScopes...))
}

func (c *container) memoryUsageMetric(stats []docker.Stats) report.Metric {
	result := report.MakeMetric()
	for _, s := range stats {
		result = result.Add(s.Read, float64(s.MemoryStats.Usage)).WithMax(float64(s.MemoryStats.Limit))
	}
	return result
}

func (c *container) cpuPercentMetric(stats []docker.Stats) report.Metric {
	result := report.MakeMetric()
	if len(stats) < 2 {
		return result
	}

	previous := stats[0]
	for _, s := range stats[1:] {
		// Copies from docker/api/client/stats.go#L205
		cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage - previous.CPUStats.CPUUsage.TotalUsage)
		systemDelta := float64(s.CPUStats.SystemCPUUsage - previous.CPUStats.SystemCPUUsage)
		cpuPercent := 0.0
		if systemDelta > 0.0 && cpuDelta > 0.0 {
			cpuPercent = (cpuDelta / systemDelta) * float64(len(s.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		}
		result = result.Add(s.Read, cpuPercent)
		available := float64(len(s.CPUStats.CPUUsage.PercpuUsage)) * 100.0
		if available >= result.Max {
			result.Max = available
		}
		previous = s
	}
	return result
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

func (c *container) getBaseNode() report.Node {
	result := report.MakeNodeWith(report.MakeContainerNodeID(c.ID()), map[string]string{
		ContainerID:       c.ID(),
		ContainerCreated:  c.container.Created.Format(time.RFC822),
		ContainerCommand:  c.container.Path + " " + strings.Join(c.container.Args, " "),
		ImageID:           c.Image(),
		ContainerHostname: c.Hostname(),
	}).WithParents(report.EmptySets.
		Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(c.Image()))),
	)
	result = result.AddTable(LabelPrefix, c.container.Config.Labels)
	result = result.AddTable(EnvPrefix, c.env())
	return result
}

func (c *container) GetNode() report.Node {
	c.RLock()
	defer c.RUnlock()
	latest := map[string]string{
		ContainerState:      c.StateString(),
		ContainerStateHuman: c.State(),
	}
	controls := []string{}

	if c.container.State.Paused {
		controls = append(controls, UnpauseContainer)
	} else if c.container.State.Running {
		uptime := (mtime.Now().Sub(c.container.State.StartedAt) / time.Second) * time.Second
		networkMode := ""
		if c.container.HostConfig != nil {
			networkMode = c.container.HostConfig.NetworkMode
		}
		latest[ContainerName] = strings.TrimPrefix(c.container.Name, "/")
		latest[ContainerUptime] = uptime.String()
		latest[ContainerRestartCount] = strconv.Itoa(c.container.RestartCount)
		latest[ContainerNetworkMode] = networkMode
		controls = append(controls, RestartContainer, StopContainer, PauseContainer, AttachContainer, ExecContainer)
	} else {
		controls = append(controls, StartContainer, RemoveContainer)
	}

	result := c.baseNode.WithLatests(latest)
	result = result.WithControls(controls...)
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
