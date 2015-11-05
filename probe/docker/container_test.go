package docker_test

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"reflect"
	"testing"
	"time"

	client "github.com/fsouza/go-dockerclient"

	"github.com/weaveworks/scope/probe/docker"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
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

	// Send some stats to the docker container
	stats := &client.Stats{}
	stats.MemoryStats.Usage = 12345
	if err = json.NewEncoder(writer).Encode(&stats); err != nil {
		t.Error(err)
	}

	// Now see if we go them
	want := report.MakeNode().WithMetadata(map[string]string{
		"docker_container_command": " ",
		"docker_container_created": "01 Jan 01 00:00 UTC",
		"docker_container_id":      "ping",
		"docker_container_name":    "pong",
		"docker_image_id":          "baz",
		"docker_label_foo1":        "bar1",
		"docker_label_foo2":        "bar2",
		"memory_usage":             "12345",
	}).WithSets(report.Sets{
		"docker_container_ports":           report.MakeStringSet("1.2.3.4:80->80/tcp", "81/tcp"),
		"docker_container_ips":             report.MakeStringSet("1.2.3.4"),
		"docker_container_ips_with_scopes": report.MakeStringSet("scope;1.2.3.4"),
	}).WithControls(
		docker.RestartContainer, docker.StopContainer, docker.PauseContainer,
	).WithLatest("docker_container_state", "running")

	test.Poll(t, 100*time.Millisecond, want, func() interface{} {
		node := c.GetNode("scope", []net.IP{})
		for k, v := range node.Metadata {
			if v == "0" || v == "" {
				delete(node.Metadata, k)
			}
		}
		return node
	})

	if c.Image() != "baz" {
		t.Errorf("%s != baz", c.Image())
	}
	if c.PID() != 1 {
		t.Errorf("%d != 1", c.PID())
	}
	if have := docker.ExtractContainerIPs(c.GetNode("", []net.IP{})); !reflect.DeepEqual(have, []string{"1.2.3.4"}) {
		t.Errorf("%v != %v", have, []string{"1.2.3.4"})
	}
}
