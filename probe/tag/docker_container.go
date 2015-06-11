package tag

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"

	docker "github.com/fsouza/go-dockerclient"
)

// These constants are keys used in node metadata
// TODO: use these constants in report/{mapping.go, detailed_node.go} - pending some circular references
const (
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

type dockerContainer struct {
	sync.RWMutex
	*docker.Container

	statsConn   *httputil.ClientConn
	latestStats *docker.Stats
}

// called whilst holding t.Lock() for writes
func (c *dockerContainer) startGatheringStats(containerID string) error {
	if c.statsConn != nil {
		return fmt.Errorf("already gather stats for container %s", containerID)
	}

	log.Printf("docker mapper: collecting stats for %s", containerID)
	req, err := http.NewRequest("GET", fmt.Sprintf("/containers/%s/stats", containerID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "weavescope")

	url, err := url.Parse(endpoint)
	if err != nil {
		return err
	}

	dial, err := net.Dial(url.Scheme, url.Path)
	if err != nil {
		return err
	}

	conn := httputil.NewClientConn(dial, nil)
	resp, err := conn.Do(req)
	if err != nil {
		return err
	}

	c.statsConn = conn

	go func() {
		defer func() {
			c.Lock()
			defer c.Unlock()

			log.Printf("docker mapper: stopped collecting stats for %s", containerID)
			c.statsConn = nil
			c.latestStats = nil
		}()

		stats := &docker.Stats{}
		decoder := json.NewDecoder(resp.Body)

		for err := decoder.Decode(&stats); err != io.EOF; err = decoder.Decode(&stats) {
			if err != nil {
				log.Printf("docker mapper: error reading event %v", err)
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

// called whilst holding t.Lock()
func (c *dockerContainer) stopGatheringStats(containerID string) {
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

// called whilst holding t.RLock()
func (c *dockerContainer) getStats() map[string]string {
	c.RLock()
	defer c.RUnlock()

	if c.latestStats == nil {
		return map[string]string{}
	}

	return map[string]string{
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
	}
}
