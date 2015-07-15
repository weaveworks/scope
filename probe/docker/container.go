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

	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
const (
	ContainerName    = "docker_container_name"
	ContainerCommand = "docker_container_command"
	ContainerPorts   = "docker_container_ports"
	ContainerCreated = "docker_container_created"

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
	ID() string
	Image() string
	PID() int
	GetNodeMetadata() report.NodeMetadata

	StartGatheringStats() error
	StopGatheringStats()
}

type container struct {
	sync.RWMutex
	container   *docker.Container
	statsConn   ClientConn
	latestStats *docker.Stats
}

// NewContainer creates a new Container
func NewContainer(c *docker.Container) Container {
	return &container{container: c}
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

		dial, err := net.Dial(url.Scheme, url.Path)
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
				log.Printf("docker container: error reading event %v", err)
				return
			}

			c.Lock()
			c.latestStats = stats
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

func (c *container) ports() string {
	if c.container.NetworkSettings == nil {
		return ""
	}

	ports := []string{}
	for port, bindings := range c.container.NetworkSettings.Ports {
		if len(bindings) == 0 {
			ports = append(ports, fmt.Sprintf("%s", port))
			continue
		}
		for _, b := range bindings {
			ports = append(ports, fmt.Sprintf("%s:%s->%s", b.HostIP, b.HostPort, port))
		}
	}

	return strings.Join(ports, ", ")
}

func (c *container) GetNodeMetadata() report.NodeMetadata {
	c.RLock()
	defer c.RUnlock()

	result := report.NewNodeMetadata(report.Metadata{
		ContainerID:      c.ID(),
		ContainerName:    strings.TrimPrefix(c.container.Name, "/"),
		ContainerPorts:   c.ports(),
		ContainerCreated: c.container.Created.Format(time.RFC822),
		ContainerCommand: c.container.Path + " " + strings.Join(c.container.Args, " "),
		ImageID:          c.container.Image,
	})

	if c.latestStats == nil {
		return result
	}

	result.Merge(report.NewNodeMetadata(report.Metadata{
		NetworkRxDropped: strconv.FormatUint(c.latestStats.Network.RxDropped, 10),
		NetworkRxBytes:   strconv.FormatUint(c.latestStats.Network.RxBytes, 10),
		NetworkRxErrors:  strconv.FormatUint(c.latestStats.Network.RxErrors, 10),
		NetworkTxPackets: strconv.FormatUint(c.latestStats.Network.TxPackets, 10),
		NetworkTxDropped: strconv.FormatUint(c.latestStats.Network.TxDropped, 10),
		NetworkRxPackets: strconv.FormatUint(c.latestStats.Network.RxPackets, 10),
		NetworkTxErrors:  strconv.FormatUint(c.latestStats.Network.TxErrors, 10),
		NetworkTxBytes:   strconv.FormatUint(c.latestStats.Network.TxBytes, 10),

		MemoryMaxUsage: strconv.FormatUint(c.latestStats.MemoryStats.MaxUsage, 10),
		MemoryUsage:    strconv.FormatUint(c.latestStats.MemoryStats.Usage, 10),
		MemoryFailcnt:  strconv.FormatUint(c.latestStats.MemoryStats.Failcnt, 10),
		MemoryLimit:    strconv.FormatUint(c.latestStats.MemoryStats.Limit, 10),

		//		CPUPercpuUsage:       strconv.FormatUint(stats.CPUStats.CPUUsage.PercpuUsage, 10),
		CPUUsageInUsermode:   strconv.FormatUint(c.latestStats.CPUStats.CPUUsage.UsageInUsermode, 10),
		CPUTotalUsage:        strconv.FormatUint(c.latestStats.CPUStats.CPUUsage.TotalUsage, 10),
		CPUUsageInKernelmode: strconv.FormatUint(c.latestStats.CPUStats.CPUUsage.UsageInKernelmode, 10),
		CPUSystemCPUUsage:    strconv.FormatUint(c.latestStats.CPUStats.SystemCPUUsage, 10),
	}))
	return result
}
