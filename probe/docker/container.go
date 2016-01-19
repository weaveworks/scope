package docker

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	docker "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	ContainerName          = "docker_container_name"
	ContainerCommand       = "docker_container_command"
	ContainerPorts         = "docker_container_ports"
	ContainerCreated       = "docker_container_created"
	ContainerIPs           = "docker_container_ips"
	ContainerHostname      = "docker_container_hostname"
	ContainerIPsWithScopes = "docker_container_ips_with_scopes"
	ContainerState         = "docker_container_state"

	NetworkRxDropped = "network_rx_dropped"
	NetworkRxBytes   = "network_rx_bytes"
	NetworkRxErrors  = "network_rx_errors"
	NetworkTxPackets = "network_tx_packets"
	NetworkTxDropped = "network_tx_dropped"
	NetworkRxPackets = "network_rx_packets"
	NetworkTxErrors  = "network_tx_errors"
	NetworkTxBytes   = "network_tx_bytes"

	MemoryMaxUsage = "memory_max_usage"
	MemoryUsage    = "memory_usage"
	MemoryFailcnt  = "memory_failcnt"
	MemoryLimit    = "memory_limit"

	CPUPercpuUsage       = "cpu_per_cpu_usage"
	CPUUsageInUsermode   = "cpu_usage_in_usermode"
	CPUTotalUsage        = "cpu_total_usage"
	CPUUsageInKernelmode = "cpu_usage_in_kernelmode"
	CPUSystemCPUUsage    = "cpu_system_cpu_usage"

	StateRunning = "running"
	StateStopped = "stopped"
	StatePaused  = "paused"

	stopTimeout = 10
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
	GetNode(string, []net.IP) report.Node
	State() string
	HasTTY() bool
	Container() *docker.Container
	StartGatheringStats() error
	StopGatheringStats()
}

type container struct {
	sync.RWMutex
	container    *docker.Container
	statsConn    ClientConn
	latestStats  *docker.Stats
	pendingStats []*docker.Stats
}

// NewContainer creates a new Container
func NewContainer(c *docker.Container) Container {
	return &container{
		container: c,
	}
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
	return c.container.Image
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
	if c.container.State.Paused {
		return StatePaused
	} else if c.container.State.Running {
		return StateRunning
	}
	return StateStopped
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
		log.Printf("docker container: collecting stats for %s", c.container.ID)
		req, err := http.NewRequest("GET", fmt.Sprintf("/containers/%s/stats", c.container.ID), nil)
		if err != nil {
			log.Printf("docker container: %v", err)
			return
		}
		req.Header.Set("User-Agent", "weavescope")

		url, err := url.Parse(endpoint)
		if err != nil {
			log.Printf("docker container: %v", err)
			return
		}

		dial, err := DialStub(url.Scheme, url.Path)
		if err != nil {
			log.Printf("docker container: %v", err)
			return
		}

		conn := NewClientConnStub(dial, nil)
		resp, err := conn.Do(req)
		if err != nil {
			log.Printf("docker container: %v", err)
			return
		}

		c.Lock()
		c.statsConn = conn
		c.Unlock()

		defer func() {
			c.Lock()
			defer c.Unlock()

			log.Printf("docker container: stopped collecting stats for %s", c.container.ID)
			c.statsConn = nil
			c.latestStats = nil
		}()

		stats := &docker.Stats{}
		decoder := json.NewDecoder(resp.Body)

		for err := decoder.Decode(&stats); err != io.EOF; err = decoder.Decode(&stats) {
			if err != nil {
				log.Printf("docker container: error reading event, did container stop? %v", err)
				return
			}

			c.Lock()
			c.latestStats = stats
			c.pendingStats = append(c.pendingStats, stats)
			c.Unlock()

			stats = &docker.Stats{}
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
	c.latestStats = nil
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
			if b.HostIP == "0.0.0.0" {
				for _, ip := range localAddrs {
					ports = append(ports, fmt.Sprintf("%s:%s->%s", ip, b.HostPort, port))
				}
			} else {
				ports = append(ports, fmt.Sprintf("%s:%s->%s", b.HostIP, b.HostPort, port))
			}
		}
	}

	return report.MakeStringSet(ports...)
}

func (c *container) memoryUsageMetric() report.Metric {
	result := report.MakeMetric()
	for _, s := range c.pendingStats {
		result = result.Add(s.Read, float64(s.MemoryStats.Usage))
	}
	return result
}

func (c *container) cpuPercentMetric() report.Metric {
	result := report.MakeMetric()
	if len(c.pendingStats) < 2 {
		return result
	}

	previous := c.pendingStats[0]
	for _, s := range c.pendingStats[1:] {
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
	result := report.Metrics{
		MemoryUsage:   c.memoryUsageMetric(),
		CPUTotalUsage: c.cpuPercentMetric(),
	}

	// Keep the latest report to help with relative metric reporting.
	if len(c.pendingStats) > 0 {
		c.pendingStats = c.pendingStats[len(c.pendingStats)-1:]
	}
	return result
}

func (c *container) GetNode(hostID string, localAddrs []net.IP) report.Node {
	c.RLock()
	defer c.RUnlock()

	ips := append(c.container.NetworkSettings.SecondaryIPAddresses, c.container.NetworkSettings.IPAddress)
	// Treat all Docker IPs as local scoped.
	ipsWithScopes := []string{}
	for _, ip := range ips {
		ipsWithScopes = append(ipsWithScopes, report.MakeScopedAddressNodeID(hostID, ip))
	}

	state := c.State()

	result := report.MakeNodeWith(map[string]string{
		ContainerID:       c.ID(),
		ContainerName:     strings.TrimPrefix(c.container.Name, "/"),
		ContainerCreated:  c.container.Created.Format(time.RFC822),
		ContainerCommand:  c.container.Path + " " + strings.Join(c.container.Args, " "),
		ImageID:           c.container.Image,
		ContainerHostname: c.Hostname(),
	}).WithSets(report.Sets{
		ContainerPorts:         c.ports(localAddrs),
		ContainerIPs:           report.MakeStringSet(ips...),
		ContainerIPsWithScopes: report.MakeStringSet(ipsWithScopes...),
	}).WithLatest(
		ContainerState, mtime.Now(), state,
	).WithMetrics(
		c.metrics(),
	).WithParents(report.Sets{
		report.ContainerImage: report.MakeStringSet(report.MakeContainerImageNodeID(c.container.Image)),
	})

	if c.container.State.Paused {
		result = result.WithControls(UnpauseContainer)
	} else if c.container.State.Running {
		result = result.WithControls(
			RestartContainer, StopContainer, PauseContainer, AttachContainer, ExecContainer,
		)
	} else {
		result = result.WithControls(StartContainer)
	}

	AddLabels(result, c.container.Config.Labels)

	if c.latestStats == nil {
		return result
	}

	result = result.WithMetadata(map[string]string{
		MemoryMaxUsage: strconv.FormatUint(c.latestStats.MemoryStats.MaxUsage, 10),
		MemoryUsage:    strconv.FormatUint(c.latestStats.MemoryStats.Usage, 10),
		MemoryFailcnt:  strconv.FormatUint(c.latestStats.MemoryStats.Failcnt, 10),
		MemoryLimit:    strconv.FormatUint(c.latestStats.MemoryStats.Limit, 10),

		//		CPUPercpuUsage:       strconv.FormatUint(stats.CPUStats.CPUUsage.PercpuUsage, 10),
		CPUUsageInUsermode:   strconv.FormatUint(c.latestStats.CPUStats.CPUUsage.UsageInUsermode, 10),
		CPUTotalUsage:        strconv.FormatUint(c.latestStats.CPUStats.CPUUsage.TotalUsage, 10),
		CPUUsageInKernelmode: strconv.FormatUint(c.latestStats.CPUStats.CPUUsage.UsageInKernelmode, 10),
		CPUSystemCPUUsage:    strconv.FormatUint(c.latestStats.CPUStats.SystemCPUUsage, 10),
	}).WithMetrics(c.metrics())
	return result
}

// ExtractContainerIPs returns the list of container IPs given a Node from the Container topology.
func ExtractContainerIPs(nmd report.Node) []string {
	return []string(nmd.Sets[ContainerIPs])
}

// ExtractContainerIPsWithScopes returns the list of container IPs, prepended
// with scopes, given a Node from the Container topology.
func ExtractContainerIPsWithScopes(nmd report.Node) []string {
	return []string(nmd.Sets[ContainerIPsWithScopes])
}
