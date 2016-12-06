package report_test

import (
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

const (
	PID    = "pid"
	Name   = "name"
	Domain = "domain"
)

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
		"Counters": {
			a: report.Nodes{
				"1": report.MakeNode("1").WithCounters(map[string]int{
					"a": 13,
					"b": 57,
					"c": 89,
				}),
			},
			b: report.Nodes{
				"1": report.MakeNode("1").WithCounters(map[string]int{
					"a": 78,
					"b": 3,
					"d": 47,
				}),
			},
			want: report.Nodes{
				"1": report.MakeNode("1").WithCounters(map[string]int{
					"a": 91,
					"b": 60,
					"c": 89,
					"d": 47,
				}),
			},
		},
	} {
		if have := c.a.Merge(c.b); !reflect.DeepEqual(c.want, have) {
			t.Errorf("%s: %s", name, test.Diff(c.want, have))
		}
	}
}
