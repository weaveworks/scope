package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/fixture"
)

func TestOriginTable(t *testing.T) {
	if _, ok := render.OriginTable(fixture.Report, "not-found", false, false); ok {
		t.Errorf("unknown origin ID gave unexpected success")
	}
	for originID, want := range map[string]render.Table{
		fixture.ServerProcessNodeID: {
			Title:   fmt.Sprintf(`Process "apache" (%s)`, fixture.ServerPID),
			Numeric: false,
			Rank:    2,
			Rows:    []render.Row{},
		},
		fixture.ServerHostNodeID: {
			Title:   fmt.Sprintf("Host %q", fixture.ServerHostName),
			Numeric: false,
			Rank:    1,
			Rows: []render.Row{
				{"Load", "0.01 0.01 0.01", "", false},
				{"Operating system", "Linux", "", false},
			},
		},
	} {
		have, ok := render.OriginTable(fixture.Report, originID, false, false)
		if !ok {
			t.Errorf("%q: not OK", originID)
			continue
		}
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", originID, test.Diff(want, have))
		}
	}

	// Test host/container tags
	for originID, want := range map[string]render.Table{
		fixture.ServerProcessNodeID: {
			Title:   fmt.Sprintf(`Process "apache" (%s)`, fixture.ServerPID),
			Numeric: false,
			Rank:    2,
			Rows: []render.Row{
				{"Host", fixture.ServerHostID, "", false},
				{"Container ID", fixture.ServerContainerID, "", false},
			},
		},
		fixture.ServerContainerNodeID: {
			Title:   `Container "server"`,
			Numeric: false,
			Rank:    3,
			Rows: []render.Row{
				{"Host", fixture.ServerHostID, "", false},
				{"ID", fixture.ServerContainerID, "", false},
				{"Image ID", fixture.ServerContainerImageID, "", false},
				{fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), `server`, "", false},
				{`Label "foo1"`, `bar1`, "", false},
				{`Label "foo2"`, `bar2`, "", false},
				{`Label "io.kubernetes.pod.name"`, "ping/pong-b", "", false},
			},
		},
	} {
		have, ok := render.OriginTable(fixture.Report, originID, true, true)
		if !ok {
			t.Errorf("%q: not OK", originID)
			continue
		}
		if !reflect.DeepEqual(want, have) {
			t.Errorf("%q: %s", originID, test.Diff(want, have))
		}
	}
}

func TestMakeDetailedHostNode(t *testing.T) {
	renderableNode := render.HostRenderer.Render(fixture.Report)[render.MakeHostID(fixture.ClientHostID)]
	have := render.MakeDetailedNode(fixture.Report, renderableNode)
	want := render.DetailedNode{
		ID:         render.MakeHostID(fixture.ClientHostID),
		LabelMajor: "client",
		LabelMinor: "hostname.com",
		Pseudo:     false,
		Controls:   []render.ControlInstance{},
		Tables: []render.Table{
			{
				Title:   fmt.Sprintf("Host %q", fixture.ClientHostName),
				Numeric: false,
				Rank:    1,
				Rows: []render.Row{
					{
						Key:        "Load",
						ValueMajor: "0.01 0.01 0.01",
						ValueMinor: "",
					},
					{
						Key:        "Operating system",
						ValueMajor: "Linux",
						ValueMinor: "",
					},
				},
			},
			{
				Title:   "Connections",
				Numeric: false,
				Rank:    0,
				Rows: []render.Row{
					{
						Key:        "TCP connections",
						ValueMajor: "3",
						ValueMinor: "",
					},
					{
						Key:        "Client",
						ValueMajor: "Server",
						ValueMinor: "",
						Expandable: true,
					},
					{
						Key:        "10.10.10.20",
						ValueMajor: "192.168.1.1",
						ValueMinor: "",
						Expandable: true,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}

func TestMakeDetailedContainerNode(t *testing.T) {
	renderableNode := render.ContainerRenderer.Render(fixture.Report)[fixture.ServerContainerID]
	have := render.MakeDetailedNode(fixture.Report, renderableNode)
	want := render.DetailedNode{
		ID:         fixture.ServerContainerID,
		LabelMajor: "server",
		LabelMinor: fixture.ServerHostName,
		Pseudo:     false,
		Controls:   []render.ControlInstance{},
		Tables: []render.Table{
			{
				Title:   `Container Image "image/server"`,
				Numeric: false,
				Rank:    4,
				Rows: []render.Row{
					{"Image ID", fixture.ServerContainerImageID, "", false},
					{`Label "foo1"`, `bar1`, "", false},
					{`Label "foo2"`, `bar2`, "", false},
				},
			},
			{
				Title:   `Container "server"`,
				Numeric: false,
				Rank:    3,
				Rows: []render.Row{
					{"ID", fixture.ServerContainerID, "", false},
					{"Image ID", fixture.ServerContainerImageID, "", false},
					{fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), `server`, "", false},
					{`Label "foo1"`, `bar1`, "", false},
					{`Label "foo2"`, `bar2`, "", false},
					{`Label "io.kubernetes.pod.name"`, "ping/pong-b", "", false},
				},
			},
			{
				Title:   fmt.Sprintf(`Process "apache" (%s)`, fixture.ServerPID),
				Numeric: false,
				Rank:    2,
				Rows:    []render.Row{},
			},
			{
				Title:   fmt.Sprintf("Host %q", fixture.ServerHostName),
				Numeric: false,
				Rank:    1,
				Rows: []render.Row{
					{"Load", "0.01 0.01 0.01", "", false},
					{"Operating system", "Linux", "", false},
				},
			},
			{
				Title:   "Connections",
				Numeric: false,
				Rank:    0,
				Rows: []render.Row{
					{"Ingress packet rate", "105", "packets/sec", false},
					{"Ingress byte rate", "1.0", "KBps", false},
					{"Client", "Server", "", true},
					{
						fmt.Sprintf("%s:%s", fixture.UnknownClient1IP, fixture.UnknownClient1Port),
						fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", fixture.UnknownClient2IP, fixture.UnknownClient2Port),
						fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", fixture.UnknownClient3IP, fixture.UnknownClient3Port),
						fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", fixture.ClientIP, fixture.ClientPort54001),
						fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", fixture.ClientIP, fixture.ClientPort54002),
						fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", fixture.RandomClientIP, fixture.RandomClientPort),
						fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						"",
						true,
					},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, have) {
		t.Errorf("%s", test.Diff(want, have))
	}
}
