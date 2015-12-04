package linux

import "testing"

func TestVMStat(t *testing.T) {
	vmstat, err := ReadVMStat("proc/vmstat")
	if err != nil {
		t.Fatal("vmstat read fail")
	}
	_ = vmstat
	t.Logf("%+v", vmstat)
}
