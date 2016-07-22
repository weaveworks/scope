package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type report struct {
	Container topology
	Plugins   []pluginSpec
}

type topology struct {
	Nodes    map[string]node    `json:"nodes"`
	Controls map[string]control `json:"controls"`
}

type node struct {
	LatestControls map[string]controlEntry `json:"latestControls,omitempty"`
}

type controlEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Value     controlData `json:"value"`
}

type controlData struct {
	Dead bool `json:"dead"`
}

type control struct {
	ID    string `json:"id"`
	Human string `json:"human"`
	Icon  string `json:"icon"`
	Rank  int    `json:"rank"`
}

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

// Reporter is responsible for generating reports.
type Reporter struct {
	store *Store
}

// NewReporter creates a new reporter that will generate reports based
// on containers saved in a given store.
func NewReporter(store *Store) *Reporter {
	return &Reporter{
		store: store,
	}
}

// RawReport generates and marshals the report for sending it to the
// probe.
func (r *Reporter) RawReport() ([]byte, error) {
	rpt := &report{
		Container: topology{
			Nodes:    r.getContainerNodes(),
			Controls: getTrafficControls(),
		},
		Plugins: []pluginSpec{
			{
				ID:          "traffic-control",
				Label:       "Traffic control",
				Description: "Adds traffic controls to the running Docker containers",
				Interfaces:  []string{"reporter", "controller"},
				APIVersion:  "1",
			},
		},
	}
	raw, err := json.Marshal(rpt)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the report: %v", err)
	}
	return raw, nil
}

// GetHandler gets the handler for a specific control in a specific
// node.
func (r *Reporter) GetHandler(nodeID, controlID string) (func() error, error) {
	containerID, err := nodeIDToContainerID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container ID from node ID %q: %v", nodeID, err)
	}
	container, found := r.store.Container(containerID)
	if !found {
		return nil, fmt.Errorf("container %q not found", containerID)
	}
	var handler func(pid int) error
	for _, c := range getControls() {
		if c.control.ID == controlID {
			handler = c.handler
			break
		}
	}
	if handler == nil {
		return nil, fmt.Errorf("unknown control ID %q for node ID %q", controlID, nodeID)
	}
	return func() error {
		return handler(container.PID)
	}, nil
}

// states:
// created, destroyed - don't create any node
// running, not running - create node with controls

func (r *Reporter) getContainerNodes() map[string]node {
	nodes := map[string]node{}
	timestamp := time.Now()
	r.store.ForEach(func(containerID string, container Container) {
		dead := false
		switch container.State {
		case Created, Destroyed:
			// do nothing, to prevent adding a stale node
			// to a report
		case Stopped:
			dead = true
			fallthrough
		case Running:
			nodeID := containerIDToNodeID(containerID)
			nodes[nodeID] = node{
				LatestControls: getTrafficNodeControls(timestamp, dead),
			}
		}
	})
	return nodes
}

func getTrafficNodeControls(timestamp time.Time, dead bool) map[string]controlEntry {
	controls := map[string]controlEntry{}
	entry := controlEntry{
		Timestamp: timestamp,
		Value: controlData{
			Dead: dead,
		},
	}
	for _, c := range getControls() {
		controls[c.control.ID] = entry
	}
	return controls
}

func getTrafficControls() map[string]control {
	controls := map[string]control{}
	for _, c := range getControls() {
		controls[c.control.ID] = c.control
	}
	return controls
}

type extControl struct {
	control control
	handler func(pid int) error
}

func getControls() []extControl {
	return []extControl{
		{
			control: control{
				ID:    "slow",
				Human: "Traffic speed: slow",
				Icon:  "fa-hourglass-1",
				Rank:  20,
			},
			handler: func(pid int) error {
				return DoTrafficControl(pid, 2000)
			},
		},
		{
			control: control{
				ID:    "medium",
				Human: "Traffic speed: medium",
				Icon:  "fa-hourglass-2",
				Rank:  21,
			},
			handler: func(pid int) error {
				return DoTrafficControl(pid, 300)
			},
		},
		{
			control: control{
				ID:    "fast",
				Human: "Traffic speed: fast",
				Icon:  "fa-hourglass-3",
				Rank:  22,
			},
			handler: func(pid int) error {
				return DoTrafficControl(pid, 1)
			},
		},
	}
}

const nodeSuffix = ";<container>"

func containerIDToNodeID(containerID string) string {
	return fmt.Sprintf("%s%s", containerID, nodeSuffix)
}

func nodeIDToContainerID(nodeID string) (string, error) {
	if !strings.HasSuffix(nodeID, nodeSuffix) {
		return "", fmt.Errorf("no suffix %q in node ID %q", nodeSuffix, nodeID)
	}
	return strings.TrimSuffix(nodeID, nodeSuffix), nil
}
