package weave_test

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/weaveworks/common/exec"
	"github.com/weaveworks/common/test"
	testExec "github.com/weaveworks/common/test/exec"
	"github.com/weaveworks/scope/common/weave"
)

const (
	mockHostID               = "host1"
	mockWeavePeerName        = "winnebago"
	mockWeavePeerNickName    = "winny"
	mockContainerID          = "83183a667c01"
	mockContainerMAC         = "d6:f2:5a:12:36:a8"
	mockContainerIP          = "10.0.0.123"
	mockContainerIPWithScope = ";10.0.0.123"
	mockHostname             = "hostname.weave.local"
	mockProxyAddress         = "unix:///foo/bar/weave.sock"
	mockDriverName           = "weave_mock"
)

var (
	mockResponse = fmt.Sprintf(`{
		"Router": {
			"Peers": [{
				"Name": "%s",
				"NickName": "%s"
			}]
		},
		"DNS": {
			"Entries": [{
				"ContainerID": "%s",
				"Hostname": "%s.",
				"Tombstone": 0
			}]
		},
                "Proxy": {
                        "Addresses": [
                                "%s"
                        ]
                },
                "Plugin": {
                        "DriverName": "%s"
                }
	}`, mockWeavePeerName, mockWeavePeerNickName, mockContainerID, mockHostname, mockProxyAddress, mockDriverName)
	mockIP = net.ParseIP("1.2.3.4")
)

func mockWeaveRouter(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte(mockResponse)); err != nil {
		panic(err)
	}
}

func TestStatus(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(mockWeaveRouter))
	defer s.Close()

	client := weave.NewClient(s.URL)
	status, err := client.Status()
	if err != nil {
		t.Fatal(err)
	}

	want := weave.Status{
		Router: weave.Router{
			Peers: []weave.Peer{
				{
					Name:     mockWeavePeerName,
					NickName: mockWeavePeerNickName,
				},
			},
		},
		DNS: &weave.DNS{
			Entries: []struct {
				Hostname    string
				ContainerID string
				Tombstone   int64
			}{
				{
					Hostname:    mockHostname + ".",
					ContainerID: mockContainerID,
					Tombstone:   0,
				},
			},
		},
		Proxy: &weave.Proxy{
			Addresses: []string{mockProxyAddress},
		},
		Plugin: &weave.Plugin{
			DriverName: mockDriverName,
		},
	}
	if !reflect.DeepEqual(status, want) {
		t.Fatal(test.Diff(status, want))
	}
}

type entry struct {
	containerid string
	ip          net.IP
}

func TestDNSAdd(t *testing.T) {
	mtx := sync.Mutex{}
	published := map[string]entry{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mtx.Lock()
		defer mtx.Unlock()
		parts := strings.SplitN(r.URL.Path, "/", 4)
		containerID, ip := parts[2], net.ParseIP(parts[3])
		fqdn := r.FormValue("fqdn")
		published[fqdn] = entry{containerID, ip}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer s.Close()

	client := weave.NewClient(s.URL)
	err := client.AddDNSEntry(mockHostname, mockContainerID, mockIP)
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]entry{
		mockHostname: {mockContainerID, mockIP},
	}
	if !reflect.DeepEqual(published, want) {
		t.Fatal(test.Diff(published, want))
	}
}

func TestPS(t *testing.T) {
	oldExecCmd := exec.Command
	defer func() { exec.Command = oldExecCmd }()
	exec.Command = func(name string, args ...string) exec.Cmd {
		return testExec.NewMockCmdString(fmt.Sprintf("%s %s %s/24\n", mockContainerID, mockContainerMAC, mockContainerIP))
	}

	client := weave.NewClient("")
	entries, err := client.PS()
	if err != nil {
		t.Fatal(err)
	}

	want := map[string]weave.PSEntry{
		mockContainerID: {
			ContainerIDPrefix: mockContainerID,
			MACAddress:        mockContainerMAC,
			IPs:               []string{mockContainerIP},
		},
	}
	if !reflect.DeepEqual(entries, want) {
		t.Fatal(test.Diff(entries, want))
	}
}

func TestExpose(t *testing.T) {
	oldExecCmd := exec.Command
	defer func() { exec.Command = oldExecCmd }()

	psCalled := false
	exec.Command = func(name string, args ...string) exec.Cmd {
		if args[0] == "expose" {
			t.Fatal("Expose not expected")
			return nil
		}
		psCalled = true
		return testExec.NewMockCmdString(fmt.Sprintf("%s %s %s/24\n", mockContainerID, mockContainerMAC, mockContainerIP))

	}

	client := weave.NewClient("")
	if err := client.Expose(); err != nil {
		t.Fatal(err)
	}

	if !psCalled {
		t.Fatal("Expected a call to weave ps")
	}

	psCalled, exposeCalled := false, false
	exec.Command = func(name string, args ...string) exec.Cmd {
		if len(args) >= 2 && args[1] == "expose" {
			exposeCalled = true
			return testExec.NewMockCmdString("")
		}
		psCalled = true
		return testExec.NewMockCmdString("")
	}

	if err := client.Expose(); err != nil {
		t.Fatal(err)
	}

	if !psCalled || !exposeCalled {
		t.Fatal("Expected a call to weave ps & expose")
	}
}
