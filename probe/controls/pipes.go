package controls

import (
	"fmt"
	"math/rand"

	"github.com/weaveworks/scope/xfer"
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

// NewPipe creats a new pipe and connects it to the app.
var NewPipe = func(c PipeClient, appID string) (string, xfer.Pipe, error) {
	pipeID := fmt.Sprintf("pipe-%d", rand.Int63())
	pipe := &pipe{
		Pipe:   xfer.NewPipe(),
		appID:  appID,
		id:     pipeID,
		client: c,
	}
	if err := c.PipeConnection(appID, pipeID, pipe.Pipe); err != nil {
		return "", nil, err
	}
	return pipeID, pipe, nil
}

func (p *pipe) Close() error {
	err1 := p.Pipe.Close()
	err2 := p.client.PipeClose(p.appID, p.id)
	if err1 != nil {
		return err1
	}
	return err2
}
