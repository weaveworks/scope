package endpoint

import (
	"bufio"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	log "github.com/Sirupsen/logrus"
)

// TCPV4TracerLocation is the location of the Python script
// that delivers the eBPF messages coming from the kernel.
// The script is located inside the Docker container in which scope executes.
var TCPV4TracerLocation = "/home/weave/tcpv4tracer.py"

// An ebpfConnection represents a TCP connection
type ebpfConnection struct {
	tuple            fourTuple
	networkNamespace string
	incoming         bool
	pid              int
}

type eventTracker interface {
	handleConnection(eventType string, tuple fourTuple, pid int, networkNamespace string)
	hasDied() bool
	run()
	walkConnections(f func(ebpfConnection))
	initialize()
	isInitialized() bool
}

// nilTracker is a tracker that does nothing, and it implements the eventTracker interface.
// It is returned when the useEbpfConn flag is false.
type nilTracker struct{}

func (n nilTracker) handleConnection(_ string, _ fourTuple, _ int, _ string) {}
func (n nilTracker) hasDied() bool                                           { return true }
func (n nilTracker) run()                                                    {}
func (n nilTracker) walkConnections(f func(ebpfConnection))                  {}
func (n nilTracker) initialize()                                             {}
func (n nilTracker) isInitialized() bool                                     { return false }

// EbpfTracker contains the sets of open and closed TCP connections.
// Closed connections are kept in the `closedConnections` slice for one iteration of `walkConnections`.
type EbpfTracker struct {
	sync.Mutex
	// the eBPF script command
	cmd *exec.Cmd

	initialized bool
	dead        bool

	openConnections   map[string]ebpfConnection
	closedConnections []ebpfConnection
}

func newEbpfTracker(useEbpfConn bool) eventTracker {
	if !useEbpfConn {
		return &nilTracker{}
	}
	cmd := exec.Command(TCPV4TracerLocation)
	env := os.Environ()
	cmd.Env = append(env, "PYTHONUNBUFFERED=1")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("EbpfTracker error: %v", err)
		return nil
	}
	go logPipe("EbpfTracker stderr:", stderr)

	tracker := &EbpfTracker{
		cmd:             cmd,
		openConnections: map[string]ebpfConnection{},
	}
	log.Info("EbpfTracker started")
	go tracker.run()
	return tracker
}

func (t *EbpfTracker) handleConnection(eventType string, tuple fourTuple, pid int, networkNamespace string) {
	t.Lock()
	defer t.Unlock()

	switch eventType {
	case "connect":
		conn := ebpfConnection{
			incoming:         false,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
		t.openConnections[tuple.String()] = conn
	case "accept":
		conn := ebpfConnection{
			incoming:         true,
			tuple:            tuple,
			pid:              pid,
			networkNamespace: networkNamespace,
		}
		t.openConnections[tuple.String()] = conn
	case "close":
		if deadConn, ok := t.openConnections[tuple.String()]; ok {
			delete(t.openConnections, tuple.String())
			t.closedConnections = append(t.closedConnections, deadConn)
		} else {
			log.Errorf("EbpfTracker error: unmatched close event: %s pid=%d netns=%s", tuple.String(), pid, networkNamespace)
		}
	}

}

func (t *EbpfTracker) run() {
	stdout, err := t.cmd.StdoutPipe()
	if err != nil {
		log.Errorf("EbpfTracker error: %v", err)
		return
	}

	if err := t.cmd.Start(); err != nil {
		log.Errorf("EbpfTracker error: %v", err)
		return
	}

	defer func() {
		if err := t.cmd.Wait(); err != nil {
			log.Errorf("EbpfTracker error: %v", err)
		}

		t.Lock()
		t.dead = true
		t.Unlock()
	}()

	reader := bufio.NewReader(stdout)
	// skip first line of output table, containing the headers
	if _, err := reader.ReadString('\n'); err != nil {
		log.Errorf("EbpfTracker error: %v", err)
		return
	}

	defer log.Infof("EbpfTracker exiting")

	scn := bufio.NewScanner(reader)
	for scn.Scan() {
		txt := scn.Text()
		line := strings.Fields(txt)

		if len(line) != 7 {
			log.Errorf("error parsing line %q", txt)
			continue
		}

		eventType := line[0]

		pid, err := strconv.Atoi(line[1])
		if err != nil {
			log.Errorf("error parsing pid %q: %v", line[1], err)
			continue
		}

		sourceAddr := net.ParseIP(line[2])
		if sourceAddr == nil {
			log.Errorf("error parsing sourceAddr %q: %v", line[2], err)
			continue
		}

		destAddr := net.ParseIP(line[3])
		if destAddr == nil {
			log.Errorf("error parsing destAddr %q: %v", line[3], err)
			continue
		}

		sPort, err := strconv.ParseUint(line[4], 10, 16)
		if err != nil {
			log.Errorf("error parsing sourcePort %q: %v", line[4], err)
			continue
		}
		sourcePort := uint16(sPort)

		dPort, err := strconv.ParseUint(line[5], 10, 16)
		if err != nil {
			log.Errorf("error parsing destPort %q: %v", line[5], err)
			continue
		}
		destPort := uint16(dPort)

		networkNamespace := line[6]

		tuple := fourTuple{sourceAddr.String(), destAddr.String(), sourcePort, destPort}

		t.handleConnection(eventType, tuple, pid, networkNamespace)
	}
}

// walkConnections calls f with all open connections and connections that have come and gone
// since the last call to walkConnections
func (t *EbpfTracker) walkConnections(f func(ebpfConnection)) {
	t.Lock()
	defer t.Unlock()

	for _, connection := range t.openConnections {
		f(connection)
	}
	for _, connection := range t.closedConnections {
		f(connection)
	}
	t.closedConnections = t.closedConnections[:0]
}

func (t *EbpfTracker) hasDied() bool {
	t.Lock()
	defer t.Unlock()

	return t.dead
}

func (t *EbpfTracker) initialize() {
	t.initialized = true
}

func (t *EbpfTracker) isInitialized() bool {
	return t.initialized
}
