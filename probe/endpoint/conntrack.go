package endpoint

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/common/exec"
)

const (
	// From https://www.kernel.org/doc/Documentation/networking/nf_conntrack-sysctl.txt
	eventsPath = "sys/net/netfilter/nf_conntrack_events"

	timeWait    = "TIME_WAIT"
	tcpProto    = "tcp"
	newType     = "NEW"
	updateType  = "UPDATE"
	destroyType = "DESTROY"
)

var (
	destroyTypeB = []byte(destroyType)
	assured      = []byte("[ASSURED] ")
	unreplied    = []byte("[UNREPLIED] ")
)

type layer3 struct {
	SrcIP string
	DstIP string
}

type layer4 struct {
	SrcPort int
	DstPort int
	Proto   string
}

type meta struct {
	Layer3 layer3
	Layer4 layer4
	ID     int64
	State  string
}

type flow struct {
	Type                         string
	Original, Reply, Independent meta
}

type conntrack struct {
	Flows []flow
}

// flowWalker is something that maintains flows, and provides an accessor
// method to walk them.
type flowWalker interface {
	walkFlows(f func(f flow, active bool))
	stop()
}

type nilFlowWalker struct{}

func (n nilFlowWalker) stop()                        {}
func (n nilFlowWalker) walkFlows(f func(flow, bool)) {}

// conntrackWalker uses the conntrack command to track network connections and
// implement flowWalker.
type conntrackWalker struct {
	sync.Mutex
	cmd           exec.Cmd
	activeFlows   map[int64]flow // active flows in state != TIME_WAIT
	bufferedFlows []flow         // flows coming out of activeFlows spend 1 walk cycle here
	bufferSize    int
	args          []string
	quit          chan struct{}
}

// newConntracker creates and starts a new conntracker.
func newConntrackFlowWalker(useConntrack bool, procRoot string, bufferSize int, args ...string) flowWalker {
	if !useConntrack {
		return nilFlowWalker{}
	} else if err := IsConntrackSupported(procRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
		return nilFlowWalker{}
	}
	result := &conntrackWalker{
		activeFlows: map[int64]flow{},
		bufferSize:  bufferSize,
		args:        args,
		quit:        make(chan struct{}),
	}
	go result.loop()
	return result
}

// IsConntrackSupported returns true if conntrack is suppported by the kernel
var IsConntrackSupported = func(procRoot string) error {
	// Make sure events are enabled, the conntrack CLI doesn't verify it
	f := filepath.Join(procRoot, eventsPath)
	contents, err := ioutil.ReadFile(f)
	if err != nil {
		return err
	}
	if string(contents) == "0" {
		return fmt.Errorf("conntrack events (%s) are disabled", f)
	}
	return nil
}

func (c *conntrackWalker) loop() {
	// conntrack can sometimes fail with ENOBUFS, when there is a particularly
	// high connection rate.  In these cases just retry in a loop, so we can
	// survive the spike.  For sustained loads this degrades nicely, as we
	// read the table before starting to handle events - basically degrading to
	// polling.
	for {
		c.run()
		c.clearFlows()

		select {
		case <-time.After(time.Second):
		case <-c.quit:
			return
		}
	}
}

func (c *conntrackWalker) clearFlows() {
	c.Lock()
	defer c.Unlock()

	for _, f := range c.activeFlows {
		c.bufferedFlows = append(c.bufferedFlows, f)
	}

	c.activeFlows = map[int64]flow{}
}

func logPipe(prefix string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Error(prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Error(prefix, err)
	}
}

func (c *conntrackWalker) run() {
	// Fork another conntrack, just to capture existing connections
	// for which we don't get events
	existingFlows, err := existingConnections(c.args)
	if err != nil {
		log.Errorf("conntrack existingConnections error: %v", err)
		return
	}
	for _, flow := range existingFlows {
		c.handleFlow(flow, true)
	}

	args := append([]string{
		"--buffer-size", strconv.Itoa(c.bufferSize), "-E",
		"-o", "id", "-p", "tcp"}, c.args...,
	)
	cmd := exec.Command("conntrack", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("conntrack error: %v", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("conntrack error: %v", err)
		return
	}
	go logPipe("conntrack stderr:", stderr)

	if err := cmd.Start(); err != nil {
		log.Errorf("conntrack error: %v", err)
		return
	}

	defer func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("conntrack error: %v", err)
		}
	}()

	c.Lock()
	// We may have stopped in the mean time,
	// so check to see if the channel is open
	// under the lock.
	select {
	default:
	case <-c.quit:
		return
	}
	c.cmd = cmd
	c.Unlock()

	scanner := NewScanner(bufio.NewReader(stdout))
	defer log.Infof("conntrack exiting")

	// Loop on the output stream
	for {
		f, err := decodeStreamedFlow(scanner)
		if err != nil {
			log.Errorf("conntrack error: %v", err)
			return
		}
		c.handleFlow(f, false)
	}
}

func existingConnections(conntrackWalkerArgs []string) ([]flow, error) {
	args := append([]string{"-L", "-o", "id", "-p", "tcp"}, conntrackWalkerArgs...)
	cmd := exec.Command("conntrack", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return []flow{}, err
	}
	if err := cmd.Start(); err != nil {
		return []flow{}, err
	}
	defer func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("conntrack existingConnections exit error: %v", err)
		}
	}()

	scanner := NewScanner(bufio.NewReader(stdout))
	var result []flow
	for {
		f, err := decodeDumpedFlow(scanner)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("conntrack error: %v", err)
			return result, err
		}
		result = append(result, f)
	}
	return result, nil
}

func (c *conntrackWalker) stop() {
	c.Lock()
	defer c.Unlock()
	close(c.quit)
	if c.cmd != nil {
		c.cmd.Kill()
	}
}

func (c *conntrackWalker) handleFlow(f flow, forceAdd bool) {
	c.Lock()
	defer c.Unlock()

	// For not, I'm only interested in tcp connections - there is too much udp
	// traffic going on (every container talking to weave dns, for example) to
	// render nicely. TODO: revisit this.
	if f.Original.Layer4.Proto != tcpProto {
		return
	}

	// Ignore flows for which we never saw an update; they are likely
	// incomplete or wrong.  See #1462.
	switch {
	case forceAdd || f.Type == updateType:
		if f.Independent.State != timeWait {
			c.activeFlows[f.Independent.ID] = f
		} else if _, ok := c.activeFlows[f.Independent.ID]; ok {
			delete(c.activeFlows, f.Independent.ID)
			c.bufferedFlows = append(c.bufferedFlows, f)
		}
	case f.Type == destroyType:
		if active, ok := c.activeFlows[f.Independent.ID]; ok {
			delete(c.activeFlows, f.Independent.ID)
			c.bufferedFlows = append(c.bufferedFlows, active)
		}
	}
}

// walkFlows calls f with all active flows and flows that have come and gone
// since the last call to walkFlows
func (c *conntrackWalker) walkFlows(f func(flow, bool)) {
	c.Lock()
	defer c.Unlock()
	for _, flow := range c.activeFlows {
		f(flow, true)
	}
	for _, flow := range c.bufferedFlows {
		f(flow, false)
	}
	c.bufferedFlows = c.bufferedFlows[:0]
}
