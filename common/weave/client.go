package weave

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/ugorji/go/codec"

	"github.com/weaveworks/scope/common/exec"
)

// Client for Weave Net API
type Client interface {
	Status() (Status, error)
	AddDNSEntry(fqdn, containerid string, ip net.IP) error
	PS() (map[string]PSEntry, error) // on the interface for mocking
	Expose() error                   // on the interface for mocking
}

// Status describes whats happen in the Weave Net router.
type Status struct {
	Router Router
	DNS    DNS
}

// Router describes the status of the Weave Router
type Router struct {
	Peers []struct {
		Name     string
		NickName string
	}
}

// DNS descirbes the status of Weave DNS
type DNS struct {
	Entries []struct {
		Hostname    string
		ContainerID string
		Tombstone   int64
	}
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
		return Status{}, err
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
		return err
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
	cmd := exec.Command("weave", "--local", "ps")
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer func() {
		if err := cmd.Wait(); err != nil {
			log.Errorf("'weave ps' cmd failed: %v", err)
		}
	}()

	psEntriesByPrefix := map[string]PSEntry{}
	scanner := bufio.NewScanner(out)
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
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return psEntriesByPrefix, nil
}

func (c *client) Expose() error {
	output, err := exec.Command("weave", "--local", "ps", "weave:expose").Output()
	if err != nil {
		return err
	}
	ips := ipMatch.FindAllSubmatch(output, -1)
	if ips != nil {
		// Alread exposed!
		return nil
	}
	if err := exec.Command("weave", "expose").Run(); err != nil {
		return fmt.Errorf("Error running weave expose: %v", err)
	}
	return nil
}
