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
	SourcePort    int
	DestPort      int
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

		sAddr := net.ParseIP(line[2])
		if sAddr == nil {
			log.Errorf("error parsing sAddr %q: %v", line[3], err)
			continue
		}

		dAddr := net.ParseIP(line[3])
		if sAddr == nil {
			log.Errorf("error parsing dAddr %q: %v", line[4], err)
			continue
		}

		sPort, err := strconv.Atoi(line[4])
		if err != nil {
			log.Errorf("error parsing sPort %q: %v", line[5], err)
			continue
		}

		dPort, err := strconv.Atoi(line[5])
		if err != nil {
			log.Errorf("error parsing dPort %q: %v", line[6], err)
			continue
		}

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
			SourceAddress: sAddr,
			DestAddress:   dAddr,
			SourcePort:    sPort,
			DestPort:      dPort,
		}

		t.Events = append(t.Events, e)
	}
}

// WalkEvents - walk through the connectionEvents
func (t *EbpfTracker) WalkEvents(f func(ConnectionEvent)) {
	for _, event := range t.Events {
		f(event)
	}
}
