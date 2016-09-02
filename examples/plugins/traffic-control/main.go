package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

// TODO:
//
// do not try to install the qdics on the network interface every
// time, skip this step if it is already installed (currently we do
// "replace" instead of "add", but this check may be a way of avoiding
// more one-time installation steps in future).
//
// somehow inform the user about the current traffic control state
// (either add some metadata about latency or maybe add background
// color to buttons to denote whether the button is active); this may
// involve sending shortcut reports as a part of a response to the
// control request
//
// detect if ip and tc binaries are in $PATH
//
// detect if required sch_netem kernel module is loaded; note that in
// some (rare) cases this might be compiled in the kernel instead of
// being a separate module; probably check if tc works, if it does not
// return something like "not implemented".
//
// add traffic control on ingress traffic too (ifb kernel module will
// be required)
//
// currently we can control latency, add controls for packet loss and
// bandwidth
//
// port to eBPF?

type containerClient interface {
	Start()
}

// Plugin is the internal data structure
type Plugin struct {
	reporter *Reporter

	clients []containerClient
}

type trafficControlStatus struct {
	latency    string
	packetLoss string
}

// String is useful to easily create a string of the traffic control plugin internal status.
// Useful for debugging
func (tcs trafficControlStatus) String() string {
	return fmt.Sprintf("%s %s", tcs.latency, tcs.packetLoss)
}

var trafficControlStatusCache map[string]trafficControlStatus
var emptyTrafficControlStatus trafficControlStatus

func main() {
	const socket = "/var/run/scope/plugins/traffic-control.sock"

	// Handle the exit signal
	setupSignals(socket)

	listener, err := setupSocket(socket)
	if err != nil {
		log.Fatalf("Failed to setup socket: %v", err)
	}

	plugin, err := NewPlugin()
	if err != nil {
		log.Fatalf("Failed to create a plugin: %v", err)
	}
	trafficControlStatusCache = make(map[string]trafficControlStatus)
	emptyTrafficControlStatus = trafficControlStatus{
		latency:    "-",
		packetLoss: "-",
	}
	if err := plugin.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func setupSignals(socket string) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	go func() {
		<-interrupt
		os.Remove(socket)
		os.Exit(0)
	}()
}

func setupSocket(socket string) (net.Listener, error) {
	os.Remove(socket)
	if err := os.MkdirAll(filepath.Dir(socket), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %q: %v", filepath.Dir(socket), err)
	}
	listener, err := net.Listen("unix", socket)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %q: %v", socket, err)
	}

	log.Printf("Listening on: unix://%s", socket)
	return listener, nil
}

// NewPlugin instantiates a new plugin
func NewPlugin() (*Plugin, error) {
	store := NewStore()
	dockerClient, err := NewDockerClient(store)
	if err != nil {
		return nil, fmt.Errorf("failed to create a docker client: %v", err)
	}
	reporter := NewReporter(store)
	plugin := &Plugin{
		reporter: reporter,
		clients: []containerClient{
			dockerClient,
		},
	}
	for _, client := range plugin.clients {
		go client.Start()
	}
	return plugin, nil
}

// Serve is a wrapper to http.ServeMux to serve the request supported by the plugin
func (p *Plugin) Serve(listener net.Listener) error {
	http.HandleFunc("/report", p.report)
	http.HandleFunc("/control", p.control)
	return http.Serve(listener, nil)
}

func (p *Plugin) report(w http.ResponseWriter, r *http.Request) {
	raw, err := p.reporter.RawReport()
	if err != nil {
		msg := fmt.Sprintf("error: failed to get raw report: %v", err)
		log.Print(msg)
		http.Error(w, msg, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}

type request struct {
	NodeID  string
	Control string
}

type response struct {
	Error string `json:"error,omitempty"`
}

func (p *Plugin) control(w http.ResponseWriter, r *http.Request) {
	xreq := request{}
	if err := json.NewDecoder(r.Body).Decode(&xreq); err != nil {
		log.Printf("Bad request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	handler, err := p.reporter.GetHandler(xreq.NodeID, xreq.Control)
	if err != nil {
		sendResponse(w, fmt.Errorf("failed to get handler: %v", err))
		return
	}
	if err := handler(); err != nil {
		sendResponse(w, fmt.Errorf("handler failed: %v", err))
		return
	}
	sendResponse(w, nil)
}

func sendResponse(w http.ResponseWriter, err error) {
	res := response{}
	if err != nil {
		res.Error = err.Error()
	}
	raw, err := json.Marshal(res)
	if err != nil {
		log.Printf("Internal server error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(raw)
}
