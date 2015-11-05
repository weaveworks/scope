package procspy

import (
	"net"
	"reflect"
	"testing"
)

func TestNetstatDarwin(t *testing.T) {
	testString := `Active Internet connections
Proto Recv-Q Send-Q  Local Address          Foreign Address        (state)
tcp4       0      0  10.0.1.6.58287         1.2.3.4.443      		ESTABLISHED
tcp4       0      0  10.0.1.6.58279         2.3.4.5.80         		ESTABLISHED
tcp4       0      0  10.0.1.6.58276         44.55.66.77.443    		ESTABLISHED
tcp4       0      0  10.0.1.6.1         	4.0.4.0.443    			GONE
`
	res := parseDarwinNetstat(testString)
	expected := []Connection{
		{
			Transport:     "tcp",
			LocalAddress:  net.ParseIP("10.0.1.6"),
			LocalPort:     58287,
			RemoteAddress: net.ParseIP("1.2.3.4"),
			RemotePort:    443,
		},
		{
			Transport:     "tcp",
			LocalAddress:  net.ParseIP("10.0.1.6"),
			LocalPort:     58279,
			RemoteAddress: net.ParseIP("2.3.4.5"),
			RemotePort:    80,
		},
		{
			Transport:     "tcp",
			LocalAddress:  net.ParseIP("10.0.1.6"),
			LocalPort:     58276,
			RemoteAddress: net.ParseIP("44.55.66.77"),
			RemotePort:    443,
		},
		/*
			{
				Transport:     "tcp",
				LocalAddress:  "::1",
				LocalPort:     "6600",
				RemoteAddress: "::1",
				RemotePort:    "41993",
			},
		*/
	}

	if len(res) != 3 {
		t.Errorf("Wanted 3")
	}
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("OS x netstat 4 error. Got\n%+v\nExpected\n%+v\n", res, expected)
	}

}
