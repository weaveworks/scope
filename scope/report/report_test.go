package report

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"
)

func TestHostJSON(t *testing.T) {
	_, localNet, _ := net.ParseCIDR("192.168.1.2/16")
	host := HostMetadata{
		Timestamp: time.Now(),
		Hostname:  "euclid",
		LocalNets: []*net.IPNet{localNet},
		OS:        "linux",
	}
	e, err := json.Marshal(host)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var hostAgain HostMetadata
	err = json.Unmarshal(e, &hostAgain)
	if err != nil {
		t.Fatalf("Unarshal error: %v", err)
	}

	// need to compare pointers. No fun.
	want := fmt.Sprintf("%+v", host)
	got := fmt.Sprintf("%+v", hostAgain)
	if want != got {
		t.Errorf("Host not the same. Want \n%+v, got \n%+v", want, got)
	}

}
