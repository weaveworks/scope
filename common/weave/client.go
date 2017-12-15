package weave

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	realexec "os/exec"
	"regexp"
	"strconv"

	"github.com/ugorji/go/codec"

	"github.com/weaveworks/common/exec"
)

const dockerAPIVersion = "1.22" // Support Docker Engine >= 1.10

// Client for Weave Net API
type Client interface {
	Status() (Status, error)
	AddDNSEntry(fqdn, containerid string, ip net.IP) error
	PS() (map[string]PSEntry, error) // on the interface for mocking
	Expose() error                   // on the interface for mocking
}

// Status describes whats happen in the Weave Net router.
type Status struct {
	Version string
	Router  Router
	DNS     *DNS
	IPAM    *IPAM
	Proxy   *Proxy
	Plugin  *Plugin
}

// Router describes the status of the Weave Router
type Router struct {
	Name               string
	Encryption         bool
	ProtocolMinVersion int
	ProtocolMaxVersion int
	PeerDiscovery      bool
	Peers              []Peer
	Connections        []struct {
		Address  string
		Outbound bool
		State    string
		Info     string
	}
	Targets        []string
	TrustedSubnets []string
}

// Peer describes a peer in the weave network
type Peer struct {
	Name        string
	NickName    string
	Connections []struct {
		Name        string
		NickName    string
		Address     string
		Outbound    bool
		Established bool
	}
}

// DNS describes the status of Weave DNS
type DNS struct {
	Domain   string
	Upstream []string
	TTL      uint32
	Entries  []struct {
		Hostname    string
		ContainerID string
		Tombstone   int64
	}
}

// IPAM describes the status of Weave IPAM
type IPAM struct {
	Paxos *struct {
		Elector    bool
		KnownNodes int
		Quorum     uint
	}
	Range         string
	DefaultSubnet string
	Entries       []struct {
		Size        uint32
		IsKnownPeer bool
	}
	PendingAllocates []string
}

// Proxy describes the status of Weave Proxy
type Proxy struct {
	Addresses []string
}

// Plugin describes the status of the Weave Plugin
type Plugin struct {
	DriverName string
}

var weavePsMatch = regexp.MustCompile(`^([0-9a-f]{12}) ((?:[0-9a-f][0-9a-f]\:){5}(?:[0-9a-f][0-9a-f]))(.*)$`)
var ipMatch = regexp.MustCompile(`([0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3})(/[0-9]+)`)

// PSEntry is a row from the output of `weave ps`
type PSEntry struct {
	ContainerIDPrefix string
	MACAddress        string
	IPs               []string
}

type client struct {
	url string
}

// NewClient makes a new Client
func NewClient(url string) Client {
	return &client{
		url: url,
	}
}

func (c *client) Status() (Status, error) {
	req, err := http.NewRequest("GET", c.url+"/report", nil)
	if err != nil {
		return Status{}, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Status{}, errorf("%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Status{}, fmt.Errorf("Got %d", resp.StatusCode)
	}

	var status Status
	decoder := codec.NewDecoder(resp.Body, &codec.JsonHandle{})
	if err := decoder.Decode(&status); err != nil {
		return Status{}, err
	}
	return status, nil
}

func (c *client) AddDNSEntry(fqdn, containerID string, ip net.IP) error {
	data := url.Values{
		"fqdn": []string{fqdn},
	}
	url := fmt.Sprintf("%s/name/%s/%s", c.url, containerID, ip.String())
	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errorf("%v", err)
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Got %d", resp.StatusCode)
	}
	return nil
}

func (c *client) PS() (map[string]PSEntry, error) {
	cmd := weaveCommand("--local", "ps")
	stdOut, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stdErr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	psEntriesByPrefix := map[string]PSEntry{}
	scanner := bufio.NewScanner(stdOut)
	for scanner.Scan() {
		line := scanner.Text()
		groups := weavePsMatch.FindStringSubmatch(line)
		if len(groups) == 0 {
			continue
		}
		containerIDPrefix, macAddress, ips := groups[1], groups[2], []string{}
		for _, ipGroup := range ipMatch.FindAllStringSubmatch(groups[3], -1) {
			ips = append(ips, ipGroup[1])
		}
		psEntriesByPrefix[containerIDPrefix] = PSEntry{
			ContainerIDPrefix: containerIDPrefix,
			MACAddress:        macAddress,
			IPs:               ips,
		}
	}
	scannerErr := scanner.Err()
	slurp, _ := ioutil.ReadAll(stdErr)
	cmdErr := cmd.Wait()
	if cmdErr != nil {
		return nil, errorf("%s: %q", cmdErr, slurp)
	}
	if scannerErr != nil {
		return nil, scannerErr
	}
	return psEntriesByPrefix, nil
}

func (c *client) Expose() error {
	cmd := weaveCommand("--local", "ps", "weave:expose")
	output, err := cmd.Output()
	if err != nil {
		stdErr := []byte{}
		if exitErr, ok := err.(*realexec.ExitError); ok {
			stdErr = exitErr.Stderr
		}
		return errorf("Error running weave ps: %s: %q", err, stdErr)
	}
	ips := ipMatch.FindAllSubmatch(output, -1)
	if ips != nil {
		// Alread exposed!
		return nil
	}
	cmd = weaveCommand("--local", "expose")
	if _, err := cmd.Output(); err != nil {
		stdErr := []byte{}
		if exitErr, ok := err.(*realexec.ExitError); ok {
			stdErr = exitErr.Stderr
		}
		return errorf("Error running weave expose: %s: %q", err, stdErr)
	}
	return nil
}

func weaveCommand(arg ...string) exec.Cmd {
	cmd := exec.Command("weave", arg...)
	cmd.SetEnv(append(os.Environ(), "DOCKER_API_VERSION="+dockerAPIVersion))
	return cmd
}

func errorf(format string, a ...interface{}) error {
	return fmt.Errorf(format+". If you are not running Weave Net, you may wish to suppress this warning by launching scope with the `--weave=false` option.", a...)
}
