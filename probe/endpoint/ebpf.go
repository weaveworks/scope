package endpoint

import (
	"bufio"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

type event int

const (
	Connect event = iota
	Accept
	Close
)

// A ConnectionEvent represents a network connection
type ConnectionEvent struct {
	Type          event
	Pid           int
	Command       string
	SourceAddress net.IP
	DestAddress   net.IP
	SourcePort    uint16
	DestPort      uint16
}

type EbpfTracker struct {
	Cmd    *exec.Cmd
	Events []ConnectionEvent
}

func NewEbpfTracker(bccProgramPath string) *EbpfTracker {
	cmd := exec.Command(bccProgramPath)
	env := os.Environ()
	cmd.Env = append(env, "PYTHONUNBUFFERED=1")

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("bcc error: %v", err)
		return nil
	}
	go logPipe("bcc stderr:", stderr)

	tracker := &EbpfTracker{
		Cmd: cmd,
	}
	go tracker.run()
	return tracker
}

func (t *EbpfTracker) run() {
	stdout, err := t.Cmd.StdoutPipe()
	if err != nil {
		log.Errorf("conntrack error: %v", err)
		return
	}

	if err := t.Cmd.Start(); err != nil {
		log.Errorf("bcc error: %v", err)
		return
	}

	defer func() {
		if err := t.Cmd.Wait(); err != nil {
			log.Errorf("bcc error: %v", err)
		}
	}()

	reader := bufio.NewReader(stdout)
	// skip fist line
	if _, err := reader.ReadString('\n'); err != nil {
		log.Errorf("bcc error: %v", err)
		return
	}

	defer log.Infof("bcc exiting")

	scn := bufio.NewScanner(reader)
	for scn.Scan() {
		txt := scn.Text()
		line := strings.Fields(txt)

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

		var evt event
		switch line[0] {
		case "connect":
			evt = Connect
		case "accept":
			evt = Accept
		case "close":
			evt = Close
		}

		e := ConnectionEvent{
			Type:          evt,
			Pid:           pid,
			SourceAddress: sourceAddr,
			DestAddress:   destAddr,
			SourcePort:    sourcePort,
			DestPort:      destPort,
		}

		t.Events = append(t.Events, e)
	}
}

// WalkEvents - walk through the connectionEvents
func (t EbpfTracker) WalkEvents(f func(ConnectionEvent)) {
	for _, event := range t.Events {
		f(event)
	}
}
