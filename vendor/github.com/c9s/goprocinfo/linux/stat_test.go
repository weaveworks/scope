package linux

import "testing"

func TestCPUStat(t *testing.T) {
	stat, err := ReadStat("proc/stat")
	if err != nil {
		t.Fatal("stat read fail")
	}
	_ = stat
	t.Logf("%+v", stat)
}
