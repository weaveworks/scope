package report_test

import (
	"bytes"
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/report"
	s_reflect "github.com/weaveworks/scope/test/reflect"
)

func TestRoundtrip(t *testing.T) {
	r1 := report.MakeReport()
	buf, _ := r1.WriteBinary()
	original := append([]byte{}, buf.Bytes()...) // copy the contents for later
	r2, err := report.MakeFromBinary(context.Background(), buf)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
	r3 := report.MakeReport()
	err = r3.ReadBinary(context.Background(), bytes.NewBuffer(original), true, true)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r1, r3) {
		t.Errorf("%v != %v", r1, r3)
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
			WithLatestControls(map[string]report.NodeControlData{"kubernetes_get_logs": {Dead: true}}).
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
	r2, err := report.MakeFromBinary(context.Background(), buf)
	if err != nil {
		t.Error(err)
	}
	if !s_reflect.DeepEqual(r1, *r2) {
		t.Errorf("%v != %v", r1, *r2)
	}
}
