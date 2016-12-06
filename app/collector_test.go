package app_test

import (
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/weaveworks/common/mtime"
	"github.com/weaveworks/scope/app"
	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test"
	"github.com/weaveworks/scope/test/reflect"
)

func TestCollector(t *testing.T) {
	ctx := context.Background()
	window := 10 * time.Second
	c := app.NewCollector(window)

	r1 := report.MakeReport()
	r1.Endpoint.AddNode(report.MakeNode("foo"))

	r2 := report.MakeReport()
	r2.Endpoint.AddNode(report.MakeNode("foo"))

	have, err := c.Report(ctx)
	if err != nil {
		t.Error(err)
	}
	if want := report.MakeReport(); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	c.Add(ctx, r1, nil)
	have, err = c.Report(ctx)
	if err != nil {
		t.Error(err)
	}
	if want := r1; !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	c.Add(ctx, r2, nil)
	merged := report.MakeReport()
	merged = merged.Merge(r1)
	merged = merged.Merge(r2)
	have, err = c.Report(ctx)
	if err != nil {
		t.Error(err)
	}
	if want := merged; !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestCollectorExpire(t *testing.T) {
	now := time.Now()
	mtime.NowForce(now)
	defer mtime.NowReset()

	ctx := context.Background()
	window := 10 * time.Second
	c := app.NewCollector(window)

	// 1st check the collector is empty
	have, err := c.Report(ctx)
	if err != nil {
		t.Error(err)
	}
	if want := report.MakeReport(); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	// Now check an added report is returned
	r1 := report.MakeReport()
	r1.Endpoint.AddNode(report.MakeNode("foo"))
	c.Add(ctx, r1, nil)
	have, err = c.Report(ctx)
	if err != nil {
		t.Error(err)
	}
	if want := r1; !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}

	// Finally move time forward to expire the report
	mtime.NowForce(now.Add(window))
	have, err = c.Report(ctx)
	if err != nil {
		t.Error(err)
	}
	if want := report.MakeReport(); !reflect.DeepEqual(want, have) {
		t.Error(test.Diff(want, have))
	}
}

func TestCollectorWait(t *testing.T) {
	ctx := context.Background()
	window := time.Millisecond
	c := app.NewCollector(window)

	waiter := make(chan struct{}, 1)
	c.WaitOn(ctx, waiter)
	defer c.UnWait(ctx, waiter)
	c.(interface {
		Broadcast()
	}).Broadcast()

	select {
	case <-waiter:
	default:
		t.Fatal("Didn't unblock")
	}
}
