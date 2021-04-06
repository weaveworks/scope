package multitenant

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"testing"

	"golang.org/x/sync/errgroup"

	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/common/xfer"
	"github.com/weaveworks/scope/probe/appclient"
)

type adapter struct {
	c appclient.AppClient
}

func (a adapter) PipeConnection(_, pipeID string, pipe xfer.Pipe) error {
	a.c.PipeConnection(pipeID, pipe)
	return nil
}

func (a adapter) PipeClose(_, pipeID string) error {
	return a.c.PipeClose(pipeID)
}

type pipeconn struct {
	id                string
	uiPR, probePR     app.PipeRouter
	uiPipe, probePipe xfer.Pipe
	uiIO, probeIO     io.ReadWriter
}

func (p *pipeconn) test(t *testing.T) {
	msg := []byte("hello " + p.id)
	wait := errgroup.Group{}

	wait.Go(func() error {
		// write something to the probe end
		_, err := p.probeIO.Write(msg)
		return err
	})

	wait.Go(func() error {
		// read it back off the other end
		buf := make([]byte, len(msg))
		n, err := p.uiIO.Read(buf)
		if err != nil {
			return err
		}
		if n != len(buf) {
			return fmt.Errorf("only read %d", n)
		}
		if !bytes.Equal(buf, msg) {
			return fmt.Errorf("Got: %v, Expected: %v", buf, msg)
		}
		return nil
	})

	err := wait.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

type pipeTest struct {
	prs   []app.PipeRouter
	pipes []*pipeconn
}

func (pt *pipeTest) newPipe(t *testing.T) {
	// make a new pipe id
	id := fmt.Sprintf("pipe-%d", rand.Int63())
	log.Printf(">>>> newPipe %s", id)

	// pick a random PR to connect app to
	uiIndex := rand.Intn(len(pt.prs))
	uiPR := pt.prs[uiIndex]
	uiPipe, uiIO, err := uiPR.Get(context.Background(), id, app.UIEnd)
	if err != nil {
		t.Fatal(err)
	}

	// pick a random PR to connect probe to
	probeIndex := rand.Intn(len(pt.prs))
	for probeIndex == uiIndex {
		probeIndex = rand.Intn(len(pt.prs))
	}
	probePR := pt.prs[probeIndex]
	probePipe, probeIO, err := probePR.Get(context.Background(), id, app.ProbeEnd)
	if err != nil {
		t.Fatal(err)
	}

	pipe := &pipeconn{
		id:        id,
		uiPR:      uiPR,
		uiPipe:    uiPipe,
		uiIO:      uiIO,
		probePR:   probePR,
		probePipe: probePipe,
		probeIO:   probeIO,
	}
	pipe.test(t)
	pt.pipes = append(pt.pipes, pipe)
}

func (pt *pipeTest) deletePipe(t *testing.T) {
	// pick a random pipe
	i := rand.Intn(len(pt.pipes))
	pipe := pt.pipes[i]
	log.Printf(">>>> deletePipe %s", pipe.id)

	if err := pipe.uiPR.Release(context.Background(), pipe.id, app.UIEnd); err != nil {
		t.Fatal(err)
	}

	if err := pipe.probePR.Release(context.Background(), pipe.id, app.ProbeEnd); err != nil {
		t.Fatal(err)
	}

	// remove from list
	pt.pipes = pt.pipes[:i+copy(pt.pipes[i:], pt.pipes[i+1:])]
}

func (pt *pipeTest) reconnectPipe(t *testing.T) {
	// pick a random pipe
	pipe := pt.pipes[rand.Intn(len(pt.pipes))]
	log.Printf(">>>> reconnectPipe %s", pipe.id)

	// pick a random PR to connect to
	newPR := pt.prs[rand.Intn(len(pt.prs))]

	// pick a random end
	if rand.Float32() < 0.5 {
		if err := pipe.uiPR.Release(context.Background(), pipe.id, app.UIEnd); err != nil {
			t.Fatal(err)
		}

		uiPipe, uiIO, err := newPR.Get(context.Background(), pipe.id, app.UIEnd)
		if err != nil {
			t.Fatal(err)
		}

		pipe.uiPR, pipe.uiPipe, pipe.uiIO = newPR, uiPipe, uiIO
	} else {
		if err := pipe.probePR.Release(context.Background(), pipe.id, app.ProbeEnd); err != nil {
			t.Fatal(err)
		}

		probePipe, probeIO, err := newPR.Get(context.Background(), pipe.id, app.ProbeEnd)
		if err != nil {
			t.Fatal(err)
		}

		pipe.probePR, pipe.probePipe, pipe.probeIO = newPR, probePipe, probeIO
	}
}

func TestPipeRouter(t *testing.T) {
	var (
		consul     = newMockConsulClient()
		replicas   = 2
		iterations = 10
		pt         = pipeTest{}
	)

	for i := 0; i < replicas; i++ {
		pr := NewConsulPipeRouter(consul, "", fmt.Sprintf("127.0.0.1:44%02d", i), NoopUserIDer)
		defer pr.Stop()
		pt.prs = append(pt.prs, pr)
	}

	for i := 0; i < iterations; i++ {
		log.Printf("Iteration %d", i)
		pt.newPipe(t)
		pt.deletePipe(t)
	}
}

//func TestPipeHard(t *testing.T) {
//	if len(pipes) <= 0 {
//		newPipe()
//		continue
//	} else if len(pipes) >= 2 {
//		deletePipe()
//		continue
//	}
//	r := rand.Float32()
//	switch {
//	case 0.0 < r && r <= 0.3:
//		newPipe()
//	case 0.3 < r && r <= 0.6:
//		deletePipe()
//	case 0.6 < r && r <= 1.0:
//		reconnectPipe()
//	}
//}
