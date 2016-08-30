package endpoint

import (
	"bufio"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/scope/common/exec"
)

const (
	// From https://www.kernel.org/doc/Documentation/networking/nf_conntrack-sysctl.txt
	// Check a tcp-related file for existence since we need tcp tracking
	procFileToCheck  = "sys/net/netfilter/nf_conntrack_tcp_timeout_close"
	xmlHeader        = "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n"
	conntrackOpenTag = "<conntrack>\n"
	timeWait         = "TIME_WAIT"
	tcpProto         = "tcp"
	newType          = "new"
	updateType       = "update"
	destroyType      = "destroy"
)

type layer3 struct {
	XMLName xml.Name `xml:"layer3"`
	SrcIP   string   `xml:"src"`
	DstIP   string   `xml:"dst"`
}

type layer4 struct {
	XMLName xml.Name `xml:"layer4"`
	SrcPort int      `xml:"sport"`
	DstPort int      `xml:"dport"`
	Proto   string   `xml:"protoname,attr"`
}

type meta struct {
	XMLName   xml.Name `xml:"meta"`
	Direction string   `xml:"direction,attr"`
	Layer3    layer3   `xml:"layer3"`
	Layer4    layer4   `xml:"layer4"`
	ID        int64    `xml:"id"`
	State     string   `xml:"state"`
}

type flow struct {
	XMLName xml.Name `xml:"flow"`
	Metas   []meta   `xml:"meta"`
	Type    string   `xml:"type,attr"`

	Original, Reply, Independent *meta `xml:"-"`
}

type conntrack struct {
	XMLName xml.Name `xml:"conntrack"`
	Flows   []flow   `xml:"flow"`
}

// flowWalker is something that maintains flows, and provides an accessor
// method to walk them.
type flowWalker interface {
	walkFlows(f func(flow))
	stop()
}

type nilFlowWalker struct{}

func (n nilFlowWalker) stop()                  {}
func (n nilFlowWalker) walkFlows(f func(flow)) {}

// conntrackWalker uses the conntrack command to track network connections and
// implement flowWalker.
type conntrackWalker struct {
	sync.Mutex
	cmd           exec.Cmd
	activeFlows   map[int64]flow // active flows in state != TIME_WAIT
	bufferedFlows []flow         // flows coming out of activeFlows spend 1 walk cycle here
	args          []string
	quit          chan struct{}
}

// newConntracker creates and starts a new conntracker.
func newConntrackFlowWalker(useConntrack bool, procRoot string, args ...string) flowWalker {
	if !useConntrack {
		return nilFlowWalker{}
	} else if err := IsConntrackSupported(procRoot); err != nil {
		log.Warnf("Not using conntrack: not supported by the kernel: %s", err)
		return nilFlowWalker{}
	}
	result := &conntrackWalker{
		activeFlows: map[int64]flow{},
		args:        args,
		quit:        make(chan struct{}),
	}
	go result.loop()
	return result
}

// IsConntrackSupported returns true if conntrack is suppported by the kernel
var IsConntrackSupported = func(procRoot string) error {
	procFile := filepath.Join(procRoot, procFileToCheck)
	_, err := os.Stat(procFile)
	return err
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
	existingFlows, err := c.existingConnections()
	if err != nil {
		log.Errorf("conntrack existingConnections error: %v", err)
		return
	}
	for _, flow := range existingFlows {
		c.handleFlow(flow, true)
	}

	args := append([]string{"-E", "-o", "xml", "-p", "tcp"}, c.args...)
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

	// Swallow the first two lines
	reader := bufio.NewReader(stdout)
	if line, err := reader.ReadString('\n'); err != nil {
		log.Errorf("conntrack error: %v", err)
		return
	} else if line != xmlHeader {
		log.Errorf("conntrack invalid output: '%s'", line)
		return
	}
	if line, err := reader.ReadString('\n'); err != nil {
		log.Errorf("conntrack error: %v", err)
		return
	} else if line != conntrackOpenTag {
		log.Errorf("conntrack invalid output: '%s'", line)
		return
	}

	defer log.Infof("contrack exiting")

	// Now loop on the output stream
	decoder := xml.NewDecoder(reader)
	for {
		var f flow
		if err := decoder.Decode(&f); err != nil {
			log.Errorf("conntrack error: %v", err)
			return
		}
		c.handleFlow(f, false)
	}
}

func (c *conntrackWalker) existingConnections() ([]flow, error) {
	args := append([]string{"-L", "-o", "xml", "-p", "tcp"}, c.args...)
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
	var result conntrack
	if err := xml.NewDecoder(stdout).Decode(&result); err == io.EOF {
		return []flow{}, nil
	} else if err != nil {
		return []flow{}, err
	}
	return result.Flows, nil
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
	// A flow consists of 3 'metas' - the 'original' 4 tuple (as seen by this
	// host) and the 'reply' 4 tuple, which is what it has been rewritten to.
	// This code finds those metas, which are identified by a Direction
	// attribute.
	for i := range f.Metas {
		meta := &f.Metas[i]
		switch meta.Direction {
		case "original":
			f.Original = meta
		case "reply":
			f.Reply = meta
		case "independent":
			f.Independent = meta
		}
	}

	// For not, I'm only interested in tcp connections - there is too much udp
	// traffic going on (every container talking to weave dns, for example) to
	// render nicely. TODO: revisit this.
	if f.Original.Layer4.Proto != tcpProto {
		return
	}

	c.Lock()
	defer c.Unlock()

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
func (c *conntrackWalker) walkFlows(f func(flow)) {
	c.Lock()
	defer c.Unlock()
	for _, flow := range c.activeFlows {
		f(flow)
	}
	for _, flow := range c.bufferedFlows {
		f(flow)
	}
	c.bufferedFlows = c.bufferedFlows[:0]
}
