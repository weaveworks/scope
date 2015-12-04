package linux

import "testing"

func TestUptime(t *testing.T) {

	uptime, err := ReadUptime("proc/uptime")
	if err != nil {
		t.Fatal(err)
	}
	if uptime.Total == 0 {
		t.Fatal("uptime total read fail")
	}
	if uptime.Idle == 0 {
		t.Fatal("uptime idel read fail")
	}

	t.Logf("Total: %+v", uptime.GetTotalDuration())
	t.Logf("Idle: %+v", uptime.GetIdleDuration())

	t.Logf("%+v", uptime)
}
