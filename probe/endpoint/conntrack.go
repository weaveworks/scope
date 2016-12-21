package endpoint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/weaveworks/common/exec"
)

const (
	// From https://www.kernel.org/doc/Documentation/networking/nf_conntrack-sysctl.txt
	// Check a tcp-related file for existence since we need tcp tracking
	procFileToCheck = "sys/net/netfilter/nf_conntrack_tcp_timeout_close"
	timeWait        = "TIME_WAIT"
	tcpProto        = "tcp"
	newType         = "[NEW]"
	updateType      = "[UPDATE]"
	destroyType     = "[DESTROY]"
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

	scanner := bufio.NewScanner(bufio.NewReader(stdout))
	defer log.Infof("conntrack exiting")

	// Lop on the output stream
	for {
		f, err := decodeStreamedFlow(scanner)
		if err != nil {
			log.Errorf("conntrack error: %v", err)
			return
		}
		c.handleFlow(f, false)
	}
}

// Get a line without [ASSURED]/[UNREPLIED] tags (it simplifies parsing)
func getUntaggedLine(scanner *bufio.Scanner) ([]byte, error) {
	success := scanner.Scan()
	if !success {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}
	line := scanner.Bytes()
	// Remove [ASSURED]/[UNREPLIED] tags
	line = removeInplace(line, assured)
	line = removeInplace(line, unreplied)
	return line, nil
}

func removeInplace(s, sep []byte) []byte {
	// TODO: See if we can get better performance
	//       removing multiple substrings at once (with index/suffixarray New()+Lookup())
	//       Probably not worth it for only two substrings occurring once.
	index := bytes.Index(s, sep)
	if index < 0 {
		return s
	}
	copy(s[index:], s[index+len(sep):])
	return s[:len(s)-len(sep)]
}

func decodeStreamedFlow(scanner *bufio.Scanner) (flow, error) {
	var (
		// Use ints for parsing unused fields since their allocations
		// are almost for free
		unused [2]int
		f      flow
	)

	// Examples:
	// " [UPDATE] udp      17 29 src=192.168.2.100 dst=192.168.2.1 sport=57767 dport=53 src=192.168.2.1 dst=192.168.2.100 sport=53 dport=57767"
	// "    [NEW] tcp      6 120 SYN_SENT src=127.0.0.1 dst=127.0.0.1 sport=58958 dport=6784 [UNREPLIED] src=127.0.0.1 dst=127.0.0.1 sport=6784 dport=58958 id=1595499776"
	// " [UPDATE] tcp      6 120 TIME_WAIT src=10.0.2.15 dst=10.0.2.15 sport=51154 dport=4040 src=10.0.2.15 dst=10.0.2.15 sport=4040 dport=51154 [ASSURED] id=3663628160"
	// " [DESTROY] tcp      6 src=172.17.0.1 dst=172.17.0.1 sport=34078 dport=53 src=172.17.0.1 dst=10.0.2.15 sport=53 dport=34078 id=3668417984" (note how the timeout and state field is missing)

	// Remove tags since they are optional and make parsing harder
	line, err := getUntaggedLine(scanner)
	if err != nil {
		return flow{}, err
	}

	line = bytes.TrimLeft(line, " ")
	if bytes.HasPrefix(line, destroyTypeB) {
		// Destroy events don't have a timeout or state field
		_, err = fmt.Sscanf(string(line), "%s %s %d src=%s dst=%s sport=%d dport=%d src=%s dst=%s sport=%d dport=%d id=%d",
			&f.Type,
			&f.Original.Layer4.Proto,
			&unused[0],
			&f.Original.Layer3.SrcIP,
			&f.Original.Layer3.DstIP,
			&f.Original.Layer4.SrcPort,
			&f.Original.Layer4.DstPort,
			&f.Reply.Layer3.SrcIP,
			&f.Reply.Layer3.DstIP,
			&f.Reply.Layer4.SrcPort,
			&f.Reply.Layer4.DstPort,
			&f.Independent.ID,
		)
	} else {
		_, err = fmt.Sscanf(string(line), "%s %s %d %d %s src=%s dst=%s sport=%d dport=%d src=%s dst=%s sport=%d dport=%d id=%d",
			&f.Type,
			&f.Original.Layer4.Proto,
			&unused[0],
			&unused[1],
			&f.Independent.State,
			&f.Original.Layer3.SrcIP,
			&f.Original.Layer3.DstIP,
			&f.Original.Layer4.SrcPort,
			&f.Original.Layer4.DstPort,
			&f.Reply.Layer3.SrcIP,
			&f.Reply.Layer3.DstIP,
			&f.Reply.Layer4.SrcPort,
			&f.Reply.Layer4.DstPort,
			&f.Independent.ID,
		)
	}

	if err != nil {
		return flow{}, fmt.Errorf("Error parsing streamed flow %q: %v ", line, err)
	}
	f.Reply.Layer4.Proto = f.Original.Layer4.Proto
	return f, nil
}

func (c *conntrackWalker) existingConnections() ([]flow, error) {
	args := append([]string{"-L", "-o", "id", "-p", "tcp"}, c.args...)
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

	scanner := bufio.NewScanner(bufio.NewReader(stdout))
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

func decodeDumpedFlow(scanner *bufio.Scanner) (flow, error) {
	var (
		// Use ints for parsing unused fields since allocations
		// are almost for free
		unused [4]int
		f      flow
	)

	// Example:
	// " tcp      6 431997 ESTABLISHED src=10.32.0.1 dst=10.32.0.1 sport=50274 dport=4040 src=10.32.0.1 dst=10.32.0.1 sport=4040 dport=50274 [ASSURED] mark=0 use=1 id=407401088c"
	// remove tags since they are optional and make parsing harder
	line, err := getUntaggedLine(scanner)
	if err != nil {
		return flow{}, err
	}

	_, err = fmt.Sscanf(string(line), "%s %d %d %s src=%s dst=%s sport=%d dport=%d src=%s dst=%s sport=%d dport=%d mark=%d use=%d id=%d",
		&f.Original.Layer4.Proto,
		&unused[0],
		&unused[1],
		&f.Independent.State,
		&f.Original.Layer3.SrcIP,
		&f.Original.Layer3.DstIP,
		&f.Original.Layer4.SrcPort,
		&f.Original.Layer4.DstPort,
		&f.Reply.Layer3.SrcIP,
		&f.Reply.Layer3.DstIP,
		&f.Reply.Layer4.SrcPort,
		&f.Reply.Layer4.DstPort,
		&unused[2],
		&unused[3],
		&f.Independent.ID,
	)

	if err != nil {
		return flow{}, fmt.Errorf("Error parsing dumped flow %q: %v ", line, err)
	}

	f.Reply.Layer4.Proto = f.Original.Layer4.Proto
	return f, nil
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
