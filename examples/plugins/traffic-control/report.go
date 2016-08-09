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
	Nodes             map[string]node             `json:"nodes"`
	Controls          map[string]control          `json:"controls"`
	MetadataTemplates map[string]metadataTemplate `json:"metadata_templates,omitempty"`
}

type metadataTemplate struct {
	ID       string  `json:"id"`
	Label    string  `json:"label,omitempty"`    // Human-readable descriptor for this row
	Truncate int     `json:"truncate,omitempty"` // If > 0, truncate the value to this length.
	Datatype string  `json:"dataType,omitempty"`
	Priority float64 `json:"priority,omitempty"`
	From     string  `json:"from,omitempty"` // Defines how to get the value from a report node
}

type node struct {
	LatestControls map[string]controlEntry `json:"latestControls,omitempty"`
	Latest         map[string]stringEntry  `json:"latest,omitempty"`
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

type stringEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Value     string    `json:"value"`
}

type pluginSpec struct {
	ID          string   `json:"id"`
	Label       string   `json:"label"`
	Description string   `json:"description,omitempty"`
	Interfaces  []string `json:"interfaces"`
	APIVersion  string   `json:"api_version,omitempty"`
}

type Reporter struct {
	store *Store
}

func NewReporter(store *Store) *Reporter {
	return &Reporter{
		store: store,
	}
}

func (r *Reporter) RawReport() ([]byte, error) {
	rpt := &report{
		Container: topology{
			Nodes:             r.getContainerNodes(),
			Controls:          getTrafficControls(),
			MetadataTemplates: getMetadataTemplate(),
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

func (r *Reporter) GetHandler(nodeID, controlID string) (func() error, error) {
	containerID, err := nodeIDToContainerID(nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get container ID from node ID %q: %v", nodeID)
	}
	container, found := r.store.Container(containerID)
	if !found {
		return nil, fmt.Errorf("container %s not found")
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
			latency, _ := getLatency(container.PID)
			nodes[nodeID] = node{
				LatestControls: getTrafficNodeControls(timestamp, dead),
				Latest: map[string]stringEntry{
					"traffic-control-latency": {
						Timestamp: timestamp,
						Value:     latency,
					},
				},
			}
		}
	})
	return nodes
}

func getMetadataTemplate() map[string]metadataTemplate {
	return map[string]metadataTemplate{
		"traffic-control-latency": metadataTemplate{
			ID:       "traffic-control-latency",
			Label:    "Latency",
			Truncate: 0,
			Datatype: "",
			Priority: 13.5,
			From:     "latest",
		},
	}
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
