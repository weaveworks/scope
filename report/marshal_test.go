package report_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/report"
	s_reflect "github.com/weaveworks/scope/test/reflect"
)

func TestRoundtrip(t *testing.T) {
	r1 := report.MakeReport()
	buf, _ := r1.WriteBinary()
	r2, err := report.MakeFromBinary(context.Background(), buf, true, true)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
}

// Create a Report for test purposes that contains about one of
// everything interesting.  Not more than one, to avoid different
// ordering comparing unequal
func makeTestReport() report.Report {
	// Note: why doesn't Scope generally force all times to UTC?
	nowTime := time.Date(2016, 12, 25, 7, 37, 0, 0, time.UTC)
	t1 := nowTime.Add(-time.Minute)
	t2 := t1.Add(time.Second)
	mtime.NowForce(nowTime)
	r := report.MakeReport()
	r.ID = "3894069658342253419"
	r.Endpoint.AddNode(report.MakeNode(";172.20.1.168;41582").
		WithTopology("endpoint").
		WithSet("snooped_dns_names", report.MakeStringSet("ip-172-20-1-168.ec2.internal")).
		WithLatestActiveControls("docker_pause_container").
		WithLatest("addr", t1, "127.0.0.1"),
	)
	r.Process.WithShape("square").WithLabel("process", "processes").
		AddNode(report.MakeNode("ip-172-20-1-168;10446").
			WithTopology("process").
			WithParents(report.MakeSets().Add("host", report.MakeStringSet("ip-172-20-1-168;<host>"))).
			WithLatest("pid", t1, "10446").
			WithMetrics(report.Metrics{"process_cpu_usage_percent": report.MakeMetric([]report.Sample{{Timestamp: t1, Value: 0.1}, {Timestamp: t2, Value: 0.2}})}))
	r.Pod.WithShape("heptagon").WithLabel("pod", "pods").
		AddNode(report.MakeNode("fceef9592ec3cf1a8e1d178fdd0de41a;<pod>").
			WithTopology("pod").
			WithLatest("host_node_id", t1, "ip-172-20-1-168;<host>"))
	r.Overlay.WithMetadataTemplates(report.MetadataTemplates{
		"weave_encryption": report.MetadataTemplate{ID: "weave_encryption", Label: "Encryption", Priority: 4, From: "latest"},
	}).
		AddNode(report.MakeNode("#docker_peer_ip-172-20-1-168").
			WithTopology("overlay").
			WithSet("local_networks", report.MakeStringSet("172.18.0.0/16")))
	mtime.NowReset()
	return r
}

func TestBiggerRoundtrip(t *testing.T) {
	r1 := makeTestReport()
	buf, _ := r1.WriteBinary()
	r2, err := report.MakeFromBinary(context.Background(), buf, true, true)
	if err != nil {
		t.Error(err)
	}
	if !s_reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
}

func TestControlsCompat(t *testing.T) {
	testData := `{
  "Container": {
    "nodes": {
      "031d;<container>": {
        "id": "031d;<container>",
        "latest": {
          "control_probe_id": {
            "timestamp": "2019-10-14T14:36:01Z",
            "value": "29b4f381044a89a3"
          }
        },
        "latestControls": {
          "docker_attach_container": {
            "timestamp": "2019-10-14T14:36:01Z",
            "value": {"dead": true}
          },
          "docker_remove_container": {
            "timestamp": "2019-10-14T14:36:01Z",
            "value": {"dead": false}
          }
        },
        "topology": "container"
      }
    },
    "shape": "hexagon"
  }
}`

	nowTime := time.Date(2019, 10, 14, 14, 36, 1, 0, time.UTC)
	mtime.NowForce(nowTime)
	expected := report.MakeReport()
	expected.Container.AddNode(report.MakeNode(report.MakeContainerNodeID("031d")).
		WithTopology("container").
		WithLatestActiveControls("docker_remove_container").
		WithLatests(map[string]string{"control_probe_id": "29b4f381044a89a3"}),
	)

	buf := bytes.NewBufferString(testData)
	rpt, err := report.MakeFromBinary(context.Background(), buf, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !s_reflect.DeepEqual(&expected, rpt) {
		t.Error(test.Diff(&expected, rpt))
	}
}
