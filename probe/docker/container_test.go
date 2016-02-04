package docker_test

import (
	"bufio"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	client "github.com/fsouza/go-dockerclient"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/common/mtime"
	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

type mockConnection struct {
	reader *io.PipeReader
}

func (c *mockConnection) Do(req *http.Request) (resp *http.Response, err error) {
	return &http.Response{
		Body: c.reader,
	}, nil
}

func (c *mockConnection) Close() error {
	return c.reader.Close()
}

func TestContainer(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	oldDialStub, oldNewClientConnStub := docker.DialStub, docker.NewClientConnStub
	defer func() { docker.DialStub, docker.NewClientConnStub = oldDialStub, oldNewClientConnStub }()

	docker.DialStub = func(network, address string) (net.Conn, error) {
		return nil, nil
	}

	reader, writer := io.Pipe()
	connection := &mockConnection{reader}

	docker.NewClientConnStub = func(c net.Conn, r *bufio.Reader) docker.ClientConn {
		return connection
	}

	c := docker.NewContainer(container1)
	err := c.StartGatheringStats()
	if err != nil {
		t.Errorf("%v", err)
	}
	defer c.StopGatheringStats()

	now := time.Unix(12345, 67890).UTC()
	mtime.NowForce(now)
	defer mtime.NowReset()

	// Send some stats to the docker container
	stats := &client.Stats{}
	stats.Read = now
	stats.MemoryStats.Usage = 12345
	encoder := codec.NewEncoder(writer, &codec.JsonHandle{})
	if err = encoder.Encode(&stats); err != nil {
		t.Error(err)
	}

	// Now see if we go them
	uptime := (now.Sub(startTime) / time.Second) * time.Second
	want := report.MakeNode().WithLatests(map[string]string{
		"docker_container_command": " ",
		"docker_container_created": "01 Jan 01 00:00 UTC",
		"docker_container_id":      "ping",
		"docker_container_name":    "pong",
		"docker_image_id":          "baz",
		"docker_label_foo1":        "bar1",
		"docker_label_foo2":        "bar2",
		"docker_container_state":   "running",
		"docker_container_uptime":  uptime.String(),
	}).WithSets(report.EmptySets.
		Add("docker_container_ports", report.MakeStringSet("1.2.3.4:80->80/tcp", "81/tcp")).
		Add("docker_container_ips", report.MakeStringSet("1.2.3.4")).
		Add("docker_container_ips_with_scopes", report.MakeStringSet("scope;1.2.3.4")),
	).WithControls(
		docker.RestartContainer, docker.StopContainer, docker.PauseContainer,
		docker.AttachContainer, docker.ExecContainer,
	).WithMetrics(report.Metrics{
		"docker_cpu_total_usage": report.MakeMetric(),
		"docker_memory_usage":    report.MakeMetric().Add(now, 12345),
	}).WithParents(report.EmptySets.
		Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID("baz"))),
	)

	test.Poll(t, 100*time.Millisecond, want, func() interface{} {
		node := c.GetNode("scope", []net.IP{})
		node.Latest.ForEach(func(k, v string) {
			if v == "0" || v == "" {
				node.Latest = node.Latest.Delete(k)
			}
		})
		return node
	})

	if c.Image() != "baz" {
		t.Errorf("%s != baz", c.Image())
	}
	if c.PID() != 2 {
		t.Errorf("%d != 2", c.PID())
	}
	if have := docker.ExtractContainerIPs(c.GetNode("", []net.IP{})); !reflect.DeepEqual(have, []string{"1.2.3.4"}) {
		t.Errorf("%v != %v", have, []string{"1.2.3.4"})
	}
}
