package app_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/weaveworks/common/test"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
)

func TestMerger(t *testing.T) {
	// Use 3 reports to check the pair-wise merging in SmartMerger
	report1 := report.MakeReport()
	report1.Endpoint.AddNode(report.MakeNode("foo"))
	report2 := report.MakeReport()
	report2.Endpoint.AddNode(report.MakeNode("bar"))
	report3 := report.MakeReport()
	report3.Endpoint.AddNode(report.MakeNode("baz"))
	reports := []report.Report{
		report1, report2, report3,
	}
	want := report.MakeReport()
	want.Endpoint.AddNode(report.MakeNode("foo"))
	want.Endpoint.AddNode(report.MakeNode("bar"))
	want.Endpoint.AddNode(report.MakeNode("baz"))

	for _, merger := range []app.Merger{app.MakeDumbMerger(), app.NewSmartMerger()} {
		// Test the empty list case
		if have := merger.Merge([]report.Report{}); !reflect.DeepEqual(have, report.MakeReport()) {
			t.Errorf("Bad merge: %s", test.Diff(have, want))
		}

		if have := merger.Merge(reports); !reflect.DeepEqual(have, want) {
			t.Errorf("Bad merge: %s", test.Diff(have, want))
		}

		// Repeat the above test to ensure caching works
		if have := merger.Merge(reports); !reflect.DeepEqual(have, want) {
			t.Errorf("Bad merge: %s", test.Diff(have, want))
		}
	}
}

func BenchmarkSmartMerger(b *testing.B) {
	benchmarkMerger(b, app.NewSmartMerger())
}

func BenchmarkDumbMerger(b *testing.B) {
	benchmarkMerger(b, app.MakeDumbMerger())
}

const numHosts = 15

func benchmarkMerger(b *testing.B, merger app.Merger) {
	makeReport := func() report.Report {
		rpt := report.MakeReport()
		for i := 0; i < 100; i++ {
			rpt.Endpoint.AddNode(report.MakeNode(fmt.Sprintf("%x", rand.Int63())))
		}
		return rpt
	}

	reports := []report.Report{}
	for i := 0; i < numHosts*5; i++ {
		reports = append(reports, makeReport())
	}
	replacements := []report.Report{}
	for i := 0; i < numHosts/3; i++ {
		replacements = append(replacements, makeReport())
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// replace 1/3 of hosts work of reports & merge them all
		for i := 0; i < len(replacements); i++ {
			reports[rand.Intn(len(reports))] = replacements[i]
		}

		merger.Merge(reports)
	}
}
