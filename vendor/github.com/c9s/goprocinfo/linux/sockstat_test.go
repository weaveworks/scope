package linux

import "testing"
import "reflect"

func TestSockStat(t *testing.T) {
	var expected = SockStat{231, 27, 1, 23, 31, 3, 19, 17, 0, 0, 0, 0}

	sockStat, err := ReadSockStat("proc/sockstat")
	if err != nil {
		t.Fatal("sockstat read fail", err)
	}

	t.Logf("%+v", sockStat)

	if !reflect.DeepEqual(*sockStat, expected) {
		t.Error("not equal to expected")
	}
}
