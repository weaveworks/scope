package render_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/weaveworks/scope/render"
	"github.com/weaveworks/scope/test"
)

func TestOriginTable(t *testing.T) {
	if _, ok := render.OriginTable(test.Report, "not-found", false, false); ok {
		t.Errorf("unknown origin ID gave unexpected success")
	}
	for originID, want := range map[string]render.Table{test.ServerProcessNodeID: {
		Title:   fmt.Sprintf(`Process "apache" (%s)`, test.ServerPID),
		Numeric: false,
		Rank:    2,
		Rows:    []render.Row{},
	},
		test.ServerHostNodeID: {
			Title:   fmt.Sprintf("Host %q", test.ServerHostName),
			Numeric: false,
			Rank:    1,
			Rows: []render.Row{
				{"Load", "0.01 0.01 0.01", "", false},
				{"Operating system", "Linux", "", false},
			},
		},
	} {
		have, ok := render.OriginTable(test.Report, originID, false, false)
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
		test.ServerProcessNodeID: {
			Title:   fmt.Sprintf(`Process "apache" (%s)`, test.ServerPID),
			Numeric: false,
			Rank:    2,
			Rows: []render.Row{
				{"Host", test.ServerHostID, "", false},
				{"Container ID", test.ServerContainerID, "", false},
			},
		},
		test.ServerContainerNodeID: {
			Title:   `Container "server"`,
			Numeric: false,
			Rank:    3,
			Rows: []render.Row{
				{"Host", test.ServerHostID, "", false},
				{"ID", test.ServerContainerID, "", false},
				{"Image ID", test.ServerContainerImageID, "", false},
				{fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), `server`, "", false},
				{`Label "foo1"`, `bar1`, "", false},
				{`Label "foo2"`, `bar2`, "", false},
				{`Label "io.kubernetes.pod.name"`, "ping/pong-b", "", false},
			},
		},
	} {
		have, ok := render.OriginTable(test.Report, originID, true, true)
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
	renderableNode := render.HostRenderer.Render(test.Report)[render.MakeHostID(test.ClientHostID)]
	have := render.MakeDetailedNode(test.Report, renderableNode)
	want := render.DetailedNode{
		ID:         render.MakeHostID(test.ClientHostID),
		LabelMajor: "client",
		LabelMinor: "hostname.com",
		Pseudo:     false,
		Tables: []render.Table{
			{
				Title:   fmt.Sprintf("Host %q", test.ClientHostName),
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
	renderableNode := render.ContainerRenderer.Render(test.Report)[test.ServerContainerID]
	have := render.MakeDetailedNode(test.Report, renderableNode)
	want := render.DetailedNode{
		ID:         test.ServerContainerID,
		LabelMajor: "server",
		LabelMinor: test.ServerHostName,
		Pseudo:     false,
		Tables: []render.Table{
			{
				Title:   `Container Image "image/server"`,
				Numeric: false,
				Rank:    4,
				Rows: []render.Row{
					{"Image ID", test.ServerContainerImageID, "", false},
					{`Label "foo1"`, `bar1`, "", false},
					{`Label "foo2"`, `bar2`, "", false},
				},
			},
			{
				Title:   `Container "server"`,
				Numeric: false,
				Rank:    3,
				Rows: []render.Row{
					{"ID", test.ServerContainerID, "", false},
					{"Image ID", test.ServerContainerImageID, "", false},
					{fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), `server`, "", false},
					{`Label "foo1"`, `bar1`, "", false},
					{`Label "foo2"`, `bar2`, "", false},
					{`Label "io.kubernetes.pod.name"`, "ping/pong-b", "", false},
				},
			},
			{
				Title:   fmt.Sprintf(`Process "apache" (%s)`, test.ServerPID),
				Numeric: false,
				Rank:    2,
				Rows:    []render.Row{},
			},
			{
				Title:   fmt.Sprintf("Host %q", test.ServerHostName),
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
						fmt.Sprintf("%s:%s", test.UnknownClient1IP, test.UnknownClient1Port),
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", test.UnknownClient2IP, test.UnknownClient2Port),
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", test.UnknownClient3IP, test.UnknownClient3Port),
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", test.ClientIP, test.ClientPort54001),
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", test.ClientIP, test.ClientPort54002),
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
						"",
						true,
					},
					{
						fmt.Sprintf("%s:%s", test.RandomClientIP, test.RandomClientPort),
						fmt.Sprintf("%s:%s", test.ServerIP, test.ServerPort),
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
