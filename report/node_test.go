package report_test

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

const (
	PID    = "pid"
	Name   = "name"
	Domain = "domain"
)

func TestWithLatest(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()

	latests1 := map[string]string{Name: "x"}
	latests2 := map[string]string{PID: "123"}
	node1 := report.MakeNode("node1").WithLatests(latests1)
	assert.Equal(t, 1, node1.Latest.Len())
	node2 := node1.WithLatests(latests1)
	assert.Equal(t, node1, node2)
	node3 := node1.WithLatests(latests2)
	assert.Equal(t, 2, node3.Latest.Len())
	node4 := node1.WithLatests(latests2)
	assert.Equal(t, node3, node4)
}

func TestMergeNodes(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()

	for name, c := range map[string]struct {
		a, b, want report.Nodes
	}{
		"Empty a": {
			a: report.Nodes{},
			b: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Empty b": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Simple merge": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{
				":192.168.1.2:12345": report.MakeNodeWith(":192.168.1.2:12345", map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
				":192.168.1.2:12345": report.MakeNodeWith(":192.168.1.2:12345", map[string]string{
					PID:    "42",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Merge conflict with rank difference": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{ // <-- same ID
					Name:   "curl",
					Domain: "node-a.local",
				}).WithLatest(PID, time.Now().Add(-1*time.Minute), "0"),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		"Merge conflict with no rank difference": {
			a: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
			b: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{ // <-- same ID
					Name:   "curl",
					Domain: "node-a.local",
				}).WithLatest(PID, time.Now().Add(-1*time.Minute), "0"),
			},
			want: report.Nodes{
				":192.168.1.1:12345": report.MakeNodeWith(":192.168.1.1:12345", map[string]string{
					PID:    "23128",
					Name:   "curl",
					Domain: "node-a.local",
				}),
			},
		},
		// Note we previously tested that counters merged by adding,
		// but that was a bug: merges must be idempotent.
		// Counters merge like other 'latest' values now.
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: %s", name, test.Diff(c.want, have))
		}
	}
}

func TestCounters(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()

	a := report.MakeNode("1").
		AddCounter("a", 13).
		AddCounter("b", 57).
		AddCounter("c", 89)

	b := a.
		AddCounter("a", 78).
		AddCounter("b", 3).
		AddCounter("d", 47)

	want := report.MakeNode("1").
		AddCounter("a", 91).
		AddCounter("b", 60).
		AddCounter("c", 89).
		AddCounter("d", 47)

	if have := b; !reflect.DeepEqual(want, have) {
		t.Errorf("Counters: %s", test.Diff(want, have))
	}
}

func TestActiveControls(t *testing.T) {
	mtime.NowForce(time.Now())
	defer mtime.NowReset()

	controls1 := []string{"bar", "foo"}
	node1 := report.MakeNode("node1").WithLatestActiveControls(controls1...)
	assert.Equal(t, controls1, node1.ActiveControls())
	assert.Equal(t, controls1, sorted(node1.MergeActiveControls(node1).ActiveControls()))

	node2 := report.MakeNode("node2")
	assert.Equal(t, controls1, node1.MergeActiveControls(node2).ActiveControls())
	assert.Equal(t, controls1, node2.MergeActiveControls(node1).ActiveControls())

	controls2 := []string{"bar", "bor"}
	controls3 := []string{"bar", "bor", "foo"}
	node3 := report.MakeNode("node1").WithLatestActiveControls(controls2...)
	assert.Equal(t, controls3, sorted(node1.MergeActiveControls(node3).ActiveControls()))
	assert.Equal(t, controls3, sorted(node3.MergeActiveControls(node1).ActiveControls()))
}

func sorted(s []string) []string {
	sort.Strings(s)
	return s
}
