package docker_test

import (
	"bufio"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"runtime"
	"testing"

	client "github.com/fsouza/go-dockerclient"
	"github.com/weaveworks/scope/probe/docker"
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
	runtime.Gosched() // wait for StartGatheringStats goroutine to call connection.Do

	// Send some stats to the docker container
	stats := &client.Stats{}
	stats.MemoryStats.Usage = 12345
	if err = json.NewEncoder(writer).Encode(&stats); err != nil {
		t.Error(err)
	}
	runtime.Gosched() // wait for StartGatheringStats goroutine to receive the stats

	// Now see if we go them
	nmd := c.GetNodeMetadata()
	if nmd[docker.MemoryUsage] != "12345" {
		t.Errorf("want 12345, got %s", nmd[docker.MemoryUsage])
	}
}
