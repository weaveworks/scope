package docker_test

import (
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

type mockStatsGatherer struct {
	opts  client.StatsOptions
	ready chan bool
}

func newMockStatsGatherer() *mockStatsGatherer {
	return &mockStatsGatherer{ready: make(chan bool)}
}

func (s *mockStatsGatherer) Stats(opts client.StatsOptions) error {
	s.opts = opts
	close(s.ready)
	return nil
}

func (s *mockStatsGatherer) Send(stats *client.Stats) {
	<-s.ready
	s.opts.Stats <- stats
}

func TestContainer(t *testing.T) {
	now := time.Unix(12345, 67890).UTC()
	mtime.NowForce(now)
	defer mtime.NowReset()

	const hostID = "scope"
	c := docker.NewContainer(container1, hostID, false, false)
	s := newMockStatsGatherer()
	err := c.StartGatheringStats(s)
	if err != nil {
		t.Errorf("%v", err)
	}
	defer c.StopGatheringStats()

	// Send some stats to the docker container
	stats := &client.Stats{}
	stats.Read = now
	stats.MemoryStats.Usage = 12345
	stats.MemoryStats.Limit = 45678
	s.Send(stats)

	// Now see if we go them
	{
		uptimeSeconds := int(now.Sub(startTime) / time.Second)
		controls := map[string]report.NodeControlData{
			docker.UnpauseContainer: {Dead: true},
			docker.RestartContainer: {Dead: false},
			docker.StopContainer:    {Dead: false},
			docker.PauseContainer:   {Dead: false},
			docker.AttachContainer:  {Dead: false},
			docker.ExecContainer:    {Dead: false},
			docker.StartContainer:   {Dead: true},
			docker.RemoveContainer:  {Dead: true},
		}
		want := report.MakeNodeWith("ping;<container>", map[string]string{
			"docker_container_command":     "ping foo.bar.local",
			"docker_container_created":     "0001-01-01T00:00:00Z",
			"docker_container_id":          "ping",
			"docker_container_name":        "pong",
			"docker_image_id":              "baz",
			"docker_label_foo1":            "bar1",
			"docker_label_foo2":            "bar2",
			"docker_container_state":       "running",
			"docker_container_state_human": c.Container().State.String(),
			"docker_container_uptime":      strconv.Itoa(uptimeSeconds),
			"docker_env_FOO":               "secret-bar",
		}).WithLatestControls(
			controls,
		).WithMetrics(report.Metrics{
			"docker_cpu_total_usage": report.MakeMetric(nil),
			"docker_memory_usage":    report.MakeSingletonMetric(now, 12345).WithMax(45678),
		}).WithParents(report.MakeSets().
			Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID("baz"))),
		)

		test.Poll(t, 100*time.Millisecond, want, func() interface{} {
			node := c.GetNode()
			latest := report.MakeStringLatestMap()
			node.Latest.ForEach(func(k string, t time.Time, v string) {
				if v != "0" && v != "" {
					latest = latest.Set(k, t, v)
				}
			})
			node.Latest = latest
			return node
		})
	}

	{
		want := report.MakeSets().
			Add("docker_container_ports", report.MakeStringSet("1.2.3.4:80->80/tcp", "81/tcp")).
			Add("docker_container_networks", nil).
			Add("docker_container_ips", report.MakeStringSet("1.2.3.4")).
			Add("docker_container_ips", report.MakeStringSet("5.6.7.8")).
			Add("docker_container_ips_with_scopes", report.MakeStringSet(";1.2.3.4")).
			Add("docker_container_ips_with_scopes", report.MakeStringSet(";5.6.7.8")).
			Add("docker_container_networks", report.MakeStringSet("network1"))

		test.Poll(t, 100*time.Millisecond, want, func() interface{} {
			return c.NetworkInfo([]net.IP{})
		})
	}

	if c.Image() != "baz" {
		t.Errorf("%s != baz", c.Image())
	}
	if c.PID() != 2 {
		t.Errorf("%d != 2", c.PID())
	}
	node := c.GetNode().WithSets(c.NetworkInfo([]net.IP{}))
	if have := docker.ExtractContainerIPs(node); !reflect.DeepEqual(have, []string{"1.2.3.4", "5.6.7.8"}) {
		t.Errorf("%v != %v", have, []string{"1.2.3.4", "5.6.7.8"})
	}
}

func TestContainerHidingArgs(t *testing.T) {
	const hostID = "scope"
	c := docker.NewContainer(container1, hostID, true, false)
	node := c.GetNode()
	node.Latest.ForEach(func(k string, _ time.Time, v string) {
		if strings.Contains(v, "foo.bar.local") {
			t.Errorf("Found command line argument in node")
		}
	})
}

func TestContainerHidingEnv(t *testing.T) {
	const hostID = "scope"
	c := docker.NewContainer(container1, hostID, false, true)
	node := c.GetNode()
	node.Latest.ForEach(func(k string, _ time.Time, v string) {
		if strings.Contains(v, "secret-bar") {
			t.Errorf("Found environment variable in node")
		}
	})
}

func TestContainerHidingBoth(t *testing.T) {
	const hostID = "scope"
	c := docker.NewContainer(container1, hostID, true, true)
	node := c.GetNode()
	node.Latest.ForEach(func(k string, _ time.Time, v string) {
		if strings.Contains(v, "foo.bar.local") {
			t.Errorf("Found command line argument in node")
		}
		if strings.Contains(v, "secret-bar") {
			t.Errorf("Found environment variable in node")
		}
	})
}
