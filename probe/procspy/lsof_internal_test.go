package procspy

import (
	"reflect"
	"testing"
)

func TestLSOFParsing(t *testing.T) {
	// List of lsof -> expected entries
	for in, expected := range map[string]map[string]Proc{
		// Single connection
		"p25196\n" +
			"ccello-app\n" +
			"n127.0.0.1:48094->127.0.0.1:4039\n" +
			"n*:4040\n": {
			"127.0.0.1:48094": {
				PID:  25196,
				Name: "cello-app",
			},
		},

		// Only listen()s.
		"cdhclient\n" +
			"n*:68\n" +
			"n*:38282\n" +
			"n*:40625\n": {},

		// A bunch
		"p13100\n" +
			"cmpd\n" +
			"n[::1]:6600\n" +
			"n127.0.0.1:6600\n" +
			"n[::1]:6600->[::1]:50992\n" +
			"p14612\n" +
			"cchromium\n" +
			"n[2003:45:2b57:8900:1869:2947:f942:aba7]:55711->[2a00:1450:4008:c01::11]:443\n" +
			"n192.168.2.111:37158->192.0.72.2:80\n" +
			"n192.168.2.111:44013->54.229.241.196:80\n" +
			"n192.168.2.111:56385->74.201.105.31:443\n" +
			"p21356\n" +
			"cssh\n" +
			"n192.168.2.111:33963->192.168.2.71:22\n": {
			"[::1]:6600": {
				PID:  13100,
				Name: "mpd",
			},
			"[2003:45:2b57:8900:1869:2947:f942:aba7]:55711": {
				PID:  14612,
				Name: "chromium",
			},
			"192.168.2.111:37158": {
				PID:  14612,
				Name: "chromium",
			},
			"192.168.2.111:44013": {
				PID:  14612,
				Name: "chromium",
			},
			"192.168.2.111:56385": {
				PID:  14612,
				Name: "chromium",
			},
			"192.168.2.111:33963": {
				PID:  21356,
				Name: "ssh",
			},
		},
	} {
		got, err := parseLSOF(in)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if !reflect.DeepEqual(expected, got) {
			t.Errorf("Expected:\n %#v\nGot:\n %#v\n", expected, got)
		}
	}
}
