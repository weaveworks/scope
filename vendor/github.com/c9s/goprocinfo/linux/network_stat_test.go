package linux

import "testing"
import "reflect"

func TestNetworkStat(t *testing.T) {

	var expected = []NetworkStat{
		{"eth0", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{"lo", 870813, 8693, 0, 0, 0, 0, 0, 0, 870813, 8693, 0, 0, 0, 0, 0, 0},
		{"virbr0", 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		{"wlan0", 1163823097, 838432, 0, 0, 0, 0, 0, 0, 73047180, 641124, 0, 0, 0, 0, 0, 0},
	}

	networkStat, err := ReadNetworkStat("proc/net_dev")
	if err != nil {
		t.Fatal("network stat read fail", err)
	}

	t.Logf("%+v", networkStat)

	if !reflect.DeepEqual(networkStat, expected) {
		t.Error("not equal to expected")
	}

	var squeezeexpected = []NetworkStat{
		{"lo", 480134461, 2323077, 0, 0, 0, 0, 0, 0, 480134461, 2323077, 0, 0, 0, 0, 0, 0},
		{"eth0", 23443246382, 63554887, 0, 0, 0, 0, 0, 0, 10900929232, 27373481, 0, 0, 0, 0, 0, 0},
	}

	networkStat, err = ReadNetworkStat("proc/net_dev_squeeze")
	if err != nil {
		t.Fatal("network stat read fail", err)
	}

	t.Logf("%+v", networkStat)

	if !reflect.DeepEqual(networkStat, squeezeexpected) {
		t.Error("not equal to expected")
	}
}
