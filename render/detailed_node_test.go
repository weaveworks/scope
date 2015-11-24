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
				{Key: "Load (1m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
				{Key: "Load (5m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
				{Key: "Load (15m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
				{Key: "Operating system", ValueMajor: "Linux"},
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
				{Key: "Host", ValueMajor: fixture.ServerHostID},
				{Key: "Container ID", ValueMajor: fixture.ServerContainerID},
			},
		},
		fixture.ServerContainerNodeID: {
			Title:   `Container "server"`,
			Numeric: false,
			Rank:    3,
			Rows: []render.Row{
				{Key: "Host", ValueMajor: fixture.ServerHostID},
				{Key: "State", ValueMajor: "running"},
				{Key: "ID", ValueMajor: fixture.ServerContainerID},
				{Key: "Image ID", ValueMajor: fixture.ServerContainerImageID},
				{Key: fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), ValueMajor: `server`},
				{Key: `Label "foo1"`, ValueMajor: `bar1`},
				{Key: `Label "foo2"`, ValueMajor: `bar2`},
				{Key: `Label "io.kubernetes.pod.name"`, ValueMajor: "ping/pong-b"},
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
		Rank:       "hostname.com",
		Pseudo:     false,
		Controls:   []render.ControlInstance{},
		Tables: []render.Table{
			{
				Title:   fmt.Sprintf("Host %q", fixture.ClientHostName),
				Numeric: false,
				Rank:    1,
				Rows: []render.Row{
					{
						Key:        "Load (1m)",
						ValueMajor: "0.01",
						Metric:     &fixture.LoadMetric,
						ValueType:  "sparkline",
					},
					{
						Key:        "Load (5m)",
						ValueMajor: "0.01",
						Metric:     &fixture.LoadMetric,
						ValueType:  "sparkline",
					},
					{
						Key:        "Load (15m)",
						ValueMajor: "0.01",
						Metric:     &fixture.LoadMetric,
						ValueType:  "sparkline",
					},
					{
						Key:        "Operating system",
						ValueMajor: "Linux",
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
					},
					{
						Key:        "Client",
						ValueMajor: "Server",
						Expandable: true,
					},
					{
						Key:        "10.10.10.20",
						ValueMajor: "192.168.1.1",
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
		Rank:       "imageid456",
		Pseudo:     false,
		Controls:   []render.ControlInstance{},
		Tables: []render.Table{
			{
				Title:   `Container Image "image/server"`,
				Numeric: false,
				Rank:    4,
				Rows: []render.Row{
					{Key: "Image ID", ValueMajor: fixture.ServerContainerImageID},
					{Key: `Label "foo1"`, ValueMajor: `bar1`},
					{Key: `Label "foo2"`, ValueMajor: `bar2`},
				},
			},
			{
				Title:   `Container "server"`,
				Numeric: false,
				Rank:    3,
				Rows: []render.Row{
					{Key: "State", ValueMajor: "running"},
					{Key: "ID", ValueMajor: fixture.ServerContainerID},
					{Key: "Image ID", ValueMajor: fixture.ServerContainerImageID},
					{Key: fmt.Sprintf(`Label %q`, render.AmazonECSContainerNameLabel), ValueMajor: `server`},
					{Key: `Label "foo1"`, ValueMajor: `bar1`},
					{Key: `Label "foo2"`, ValueMajor: `bar2`},
					{Key: `Label "io.kubernetes.pod.name"`, ValueMajor: "ping/pong-b"},
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
					{Key: "Load (1m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
					{Key: "Load (5m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
					{Key: "Load (15m)", ValueMajor: "0.01", Metric: &fixture.LoadMetric, ValueType: "sparkline"},
					{Key: "Operating system", ValueMajor: "Linux"},
				},
			},
			{
				Title:   "Connections",
				Numeric: false,
				Rank:    0,
				Rows: []render.Row{
					{Key: "Ingress packet rate", ValueMajor: "105", ValueMinor: "packets/sec"},
					{Key: "Ingress byte rate", ValueMajor: "1.0", ValueMinor: "KBps"},
					{Key: "Client", ValueMajor: "Server", Expandable: true},
					{
						Key:        fmt.Sprintf("%s:%s", fixture.UnknownClient1IP, fixture.UnknownClient1Port),
						ValueMajor: fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						Expandable: true,
					},
					{
						Key:        fmt.Sprintf("%s:%s", fixture.UnknownClient2IP, fixture.UnknownClient2Port),
						ValueMajor: fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						Expandable: true,
					},
					{
						Key:        fmt.Sprintf("%s:%s", fixture.UnknownClient3IP, fixture.UnknownClient3Port),
						ValueMajor: fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						Expandable: true,
					},
					{
						Key:        fmt.Sprintf("%s:%s", fixture.ClientIP, fixture.ClientPort54001),
						ValueMajor: fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						Expandable: true,
					},
					{
						Key:        fmt.Sprintf("%s:%s", fixture.ClientIP, fixture.ClientPort54002),
						ValueMajor: fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
						Expandable: true,
					},
					{
						Key:        fmt.Sprintf("%s:%s", fixture.RandomClientIP, fixture.RandomClientPort),
						ValueMajor: fmt.Sprintf("%s:%s", fixture.ServerIP, fixture.ServerPort),
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
