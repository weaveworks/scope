package controls

import (
	"fmt"
	"io"
	"math/rand"

	"github.com/weaveworks/scope/common/xfer"
)

// PipeClient is the type of the thing the probe uses to make pipe connections.
type PipeClient interface {
	PipeConnection(string, string, xfer.Pipe) error
	PipeClose(string, string) error
}

// pipe is the probe-local type for a pipe, extending
// xfer.Pipe with the appID and a custom closer method.
type pipe struct {
	xfer.Pipe
	id, appID string
	client    PipeClient
}

func newPipe(p xfer.Pipe, c PipeClient, appID string) (string, xfer.Pipe, error) {
	pipeID := fmt.Sprintf("pipe-%d", rand.Int63())
	pipe := &pipe{
		Pipe:   p,
		appID:  appID,
		id:     pipeID,
		client: c,
	}
	if err := c.PipeConnection(appID, pipeID, pipe.Pipe); err != nil {
		return "", nil, err
	}
	return pipeID, pipe, nil
}

// NewPipe creates a new pipe and connects it to the app.
var NewPipe = func(c PipeClient, appID string) (string, xfer.Pipe, error) {
	return newPipe(xfer.NewPipe(), c, appID)
}

// NewPipeFromEnds creates a new pipe from its ends and connects it to the app.
func NewPipeFromEnds(local, remote io.ReadWriter, c PipeClient, appID string) (string, xfer.Pipe, error) {
	return newPipe(xfer.NewPipeFromEnds(local, remote), c, appID)
}

func (p *pipe) Close() error {
	err1 := p.Pipe.Close()
	err2 := p.client.PipeClose(p.appID, p.id)
	if err1 != nil {
		return err1
	}
	return err2
}

// DummyPipeClient implements PipeClient when running the probe in debugging mode
type DummyPipeClient struct{}

// PipeConnection implements controls.PipeClient
func (DummyPipeClient) PipeConnection(appID, pipeID string, pipe xfer.Pipe) error { return nil }

// PipeClose implements controls.PipeClient
func (DummyPipeClient) PipeClose(appID, pipeID string) error { return nil }
