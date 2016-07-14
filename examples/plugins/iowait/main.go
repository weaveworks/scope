package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

func main() {
	hostname, _ := os.Hostname()
	var (
		addr   = flag.String("addr", "/var/run/scope/plugins/iowait.sock", "unix socket to listen for connections on")
		hostID = flag.String("hostname", hostname, "hostname of the host running this plugin")
	)
	flag.Parse()

	log.Printf("Starting on %s...\n", *hostID)

	// Check we can get the iowait for the system
	_, err := iowait()
	if err != nil {
		log.Fatal(err)
	}

	os.Remove(*addr)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		os.Remove(*addr)
		os.Exit(0)
	}()

	listener, err := net.Listen("unix", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.Remove(*addr)
	}()

	log.Printf("Listening on: unix://%s", *addr)

	plugin := &Plugin{HostID: *hostID}
	http.HandleFunc("/report", plugin.Report)
	http.HandleFunc("/control", plugin.Control)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

// Plugin groups the methods a plugin needs
type Plugin struct {
	HostID     string
	iowaitMode bool
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	metric, metricTemplate, err := p.metricsSnippets()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	topologyControl, nodeControl := p.controlsSnippets()
	rpt := fmt.Sprintf(`{
		 "Host": {
			 "nodes": {
				 %q: {
					 "metrics": { %s },
					 "controls": { %s }
				 }
			 },
			 "metric_templates": { %s },
			 "controls": { %s }
		 },
		 "Plugins": [
			 {
				 "id":          "iowait",
				 "label":       "iowait",
				 "description": "Adds a graph of CPU IO Wait to hosts",
				 "interfaces":  ["reporter", "controller"],
				 "api_version": "1"
			 }
		 ]
	 }`, p.getTopologyHost(), metric, nodeControl, metricTemplate, topologyControl)
	fmt.Fprintf(w, "%s", rpt)
}

// Request is just a trimmed down xfer.Request
type Request struct {
	NodeID  string
	Control string
}

// Control is called by scope when a control is activated. It is part
// of the "controller" interface.
func (p *Plugin) Control(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	xreq := Request{}
	err := json.NewDecoder(r.Body).Decode(&xreq)
	if err != nil {
		log.Printf("Bad request: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	thisNodeID := p.getTopologyHost()
	if xreq.NodeID != thisNodeID {
		log.Printf("Bad nodeID, expected %q, got %q", thisNodeID, xreq.NodeID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	expectedControlID, _, _ := p.controlDetails()
	if expectedControlID != xreq.Control {
		log.Printf("Bad control, expected %q, got %q", expectedControlID, xreq.Control)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	p.iowaitMode = !p.iowaitMode
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "{}")
}

func (p *Plugin) getTopologyHost() string {
	return fmt.Sprintf("%s;<host>", p.HostID)
}

// Get the metrics and metric_templates JSON snippets
func (p *Plugin) metricsSnippets() (string, string, error) {
	id, name := p.metricIDAndName()
	value, err := p.metricValue()
	if err != nil {
		return "", "", err
	}
	nowISO := rfcNow()
	metric := fmt.Sprintf(`
		 %q: {
			 "samples": [ {"date": %q, "value": %f} ],
			 "min": 0,
			 "max": 100
		 }
`, id, nowISO, value)
	metricTemplate := fmt.Sprintf(`
		 %q: {
			 "id":       %q,
			 "label":    %q,
			 "format":   "percent",
			 "priority": 0.1
		 }
`, id, id, name)
	return metric, metricTemplate, nil
}

// Get the topology controls and node's controls JSON snippet
func (p *Plugin) controlsSnippets() (string, string) {
	id, human, icon := p.controlDetails()
	nowISO := rfcNow()
	topologyControl := fmt.Sprintf(`
		%q: {
			"id": %q,
			"human": %q,
			"icon": %q,
			"rank": 1
		}
`, id, id, human, icon)
	nodeControl := fmt.Sprintf(`
		"timestamp": %q,
		"controls": [%q]
`, nowISO, id)
	return topologyControl, nodeControl
}

func rfcNow() string {
	now := time.Now()
	return now.Format(time.RFC3339)
}

func (p *Plugin) metricIDAndName() (string, string) {
	if p.iowaitMode {
		return "iowait", "IO Wait"
	}
	return "idle", "Idle"
}

func (p *Plugin) metricValue() (float64, error) {
	if p.iowaitMode {
		return iowait()
	}
	return idle()
}

func (p *Plugin) controlDetails() (string, string, string) {
	if p.iowaitMode {
		return "switchToIdle", "Switch to idle", "fa-beer"
	}
	return "switchToIOWait", "Switch to IO wait", "fa-hourglass"
}

// Get the latest iowait value
func iowait() (float64, error) {
	return iostatValue(3)
}

func idle() (float64, error) {
	return iostatValue(5)
}

func iostatValue(idx int) (float64, error) {
	values, err := iostat()
	if err != nil {
		return 0, err
	}
	if idx >= len(values) {
		return 0, fmt.Errorf("invalid iostat field index %d", idx)
	}

	return strconv.ParseFloat(values[idx], 64)
}

// Get the latest iostat values
func iostat() ([]string, error) {
	out, err := exec.Command("iostat", "-c").Output()
	if err != nil {
		return nil, fmt.Errorf("iowait: %v", err)
	}

	// Linux 4.2.0-25-generic (a109563eab38)	04/01/16	_x86_64_(4 CPU)
	//
	// avg-cpu:  %user   %nice %system %iowait  %steal   %idle
	//	          2.37    0.00    1.58    0.01    0.00   96.04
	lines := strings.Split(string(out), "\n")
	if len(lines) < 4 {
		return nil, fmt.Errorf("iowait: unexpected output: %q", out)
	}

	values := strings.Fields(lines[3])
	if len(values) != 6 {
		return nil, fmt.Errorf("iowait: unexpected output: %q", out)
	}
	return values, nil
}
