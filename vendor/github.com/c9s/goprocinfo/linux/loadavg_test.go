package linux

import (
	"reflect"
	"testing"
)

func TestLoadAvg(t *testing.T) {

	loadavg, err := ReadLoadAvg("proc/loadavg")

	if err != nil {
		t.Fatal("read loadavg fail", err)
	}

	expected := &LoadAvg{
		Last1Min:       0.01,
		Last5Min:       0.02,
		Last15Min:      0.05,
		ProcessRunning: 1,
		ProcessTotal:   135,
		LastPID:        11975,
	}

	if !reflect.DeepEqual(loadavg, expected) {
		t.Errorf("not equal to expected %+v", expected)
	}

	t.Logf("%+v", loadavg)
}
