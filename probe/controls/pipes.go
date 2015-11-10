package controls

import (
	"fmt"
	"math/rand"

	"github.com/weaveworks/scope/xfer"
)

// Client is the thing the probe uses to make pipe connections.
var Client interface {
	PipeConnection(string, string, xfer.Pipe) error
	PipeClose(string, string) error
}

// Pipe the probe-local type for a pipe, extending
// xfer.Pipe with the appID and a custom closer method.
type Pipe interface {
	xfer.Pipe
}

type pipe struct {
	xfer.Pipe
	id, appID string
}

// NewPipe creats a new pipe and connects it to the app.
var NewPipe = func(appID string) (string, Pipe, error) {
	pipeID := fmt.Sprintf("pipe-%d", rand.Int63())
	pipe := &pipe{
		Pipe:  xfer.NewPipe(),
		appID: appID,
		id:    pipeID,
	}
	if err := Client.PipeConnection(appID, pipeID, pipe.Pipe); err != nil {
		return "", nil, err
	}
	return pipeID, pipe, nil
}

func (p *pipe) Close() error {
	err1 := p.Pipe.Close()
	err2 := Client.PipeClose(p.appID, p.id)
	if err1 != nil {
		return err1
	}
	return err2
}
