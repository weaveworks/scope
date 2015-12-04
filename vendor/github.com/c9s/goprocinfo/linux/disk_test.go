package linux

import "testing"

func TestDisk(t *testing.T) {
	disk, err := ReadDisk("/")
	t.Logf("%+v", disk)
	if err != nil {
		t.Fatal("disk read fail")
	}
	if disk.Free <= 0 {
		t.Log("no good")
	}
}
