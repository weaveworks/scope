package main

import (
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

	log.Println("Starting...")

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
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

// Plugin groups the methods a plugin needs
type Plugin struct {
	HostID string
}

// Report is called by scope when a new report is needed. It is part of the
// "reporter" interface, which all plugins must implement.
func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.String())
	now := time.Now()
	nowISO := now.Format(time.RFC3339)
	value, err := iowait()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, `{
		 "Host": {
			 "nodes": {
				 %q: {
					 "metrics": {
						 "iowait": {
							 "samples": [ {"date": %q, "value": %f} ]
						 }
					 }
				 }
			 },
			 "metric_templates": {
				 "iowait": {
					 "id":       "iowait",
					 "label":    "IO Wait",
					 "format":   "percent",
					 "priority": 0.1
				 }
			 }
		 },
		 "Plugins": [
			 {
				 "id":          "iowait",
				 "label":       "iowait",
				 "description": "Adds a graph of CPU IO Wait to hosts",
				 "interfaces":  ["reporter"],
				 "api_version": "1"
			 }
		 ]
	 }`, p.HostID+";<host>", nowISO, value)
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// Get the latest iowait value
func iowait() (float64, error) {
	out, err := exec.Command("iostat", "-c").Output()
	if err != nil {
		return 0, fmt.Errorf("iowait: %v", err)
	}

	// Linux 4.2.0-25-generic (a109563eab38)	04/01/16	_x86_64_(4 CPU)
	//
	// avg-cpu:  %user   %nice %system %iowait  %steal   %idle
	//	          2.37    0.00    1.58    0.01    0.00   96.04
	lines := strings.Split(string(out), "\n")
	if len(lines) < 4 {
		return 0, fmt.Errorf("iowait: unexpected output: %q", out)
	}

	values := strings.Fields(lines[3])
	if len(values) != 6 {
		return 0, fmt.Errorf("iowait: unexpected output: %q", out)
	}

	return strconv.ParseFloat(values[3], 64)
}
