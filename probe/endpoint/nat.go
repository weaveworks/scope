package endpoint

import (
	"bufio"
	"encoding/xml"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/weaveworks/scope/report"
)

const (
	modules         = "/proc/modules"
	conntrackModule = "nf_conntrack"
)

// these structs are for the parsed conntrack output
type layer3 struct {
	XMLName xml.Name `xml:"layer3"`
	SrcIP   string   `xml:"src"`
	DstIP   string   `xml:"dst"`
}

type layer4 struct {
	XMLName xml.Name `xml:"layer4"`
	SrcPort int      `xml:"sport"`
	DstPort int      `xml:"dport"`
	Proto   string   `xml:"protoname,attr"`
}

type meta struct {
	XMLName   xml.Name `xml:"meta"`
	Direction string   `xml:"direction,attr"`
	Layer3    layer3   `xml:"layer3"`
	Layer4    layer4   `xml:"layer4"`
}

type flow struct {
	XMLName xml.Name `xml:"flow"`
	Metas   []meta   `xml:"meta"`
}

type conntrack struct {
	XMLName xml.Name `xml:"conntrack"`
	Flows   []flow   `xml:"flow"`
}

// This is our 'abstraction' of the endpoint that have been rewritten by NAT.
// Original is the private IP that has been rewritten.
type endpointMapping struct {
	originalIP   string
	originalPort int

	rewrittenIP   string
	rewrittenPort int
}

// natTable returns a list of endpoints that have been remapped by NAT.
func natTable() ([]endpointMapping, error) {
	var conntrack conntrack
	cmd := exec.Command("conntrack", "-L", "--any-nat", "-o", "xml")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	defer func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("conntrack error: %v", err)
		}
	}()
	if err := xml.NewDecoder(stdout).Decode(&conntrack); err != nil {
		if err == io.EOF {
			return []endpointMapping{}, nil
		}
		return nil, err
	}

	output := []endpointMapping{}
	for _, flow := range conntrack.Flows {
		// A flow consists of 3 'metas' - the 'original' 4 tuple (as seen by this
		// host) and the 'reply' 4 tuple, which is what it has been rewritten to.
		// This code finds those metas, which are identified by a Direction
		// attribute.
		original, reply := meta{}, meta{}
		for _, meta := range flow.Metas {
			if meta.Direction == "original" {
				original = meta
			} else if meta.Direction == "reply" {
				reply = meta
			}
		}

		if original.Layer4.Proto != "tcp" {
			continue
		}

		var conn endpointMapping
		if original.Layer3.SrcIP == reply.Layer3.DstIP {
			conn = endpointMapping{
				originalIP:    reply.Layer3.SrcIP,
				originalPort:  reply.Layer4.SrcPort,
				rewrittenIP:   original.Layer3.DstIP,
				rewrittenPort: original.Layer4.DstPort,
			}
		} else {
			conn = endpointMapping{
				originalIP:    original.Layer3.SrcIP,
				originalPort:  original.Layer4.SrcPort,
				rewrittenIP:   reply.Layer3.DstIP,
				rewrittenPort: reply.Layer4.DstPort,
			}
		}

		output = append(output, conn)
	}

	return output, nil
}

// applyNAT duplicates NodeMetadatas in the endpoint topology of a
// report, based on the NAT table as returns by natTable.
func applyNAT(rpt report.Report, scope string) error {
	mappings, err := natTable()
	if err != nil {
		return err
	}

	for _, mapping := range mappings {
		realEndpointID := report.MakeEndpointNodeID(scope, mapping.originalIP, strconv.Itoa(mapping.originalPort))
		copyEndpointID := report.MakeEndpointNodeID(scope, mapping.rewrittenIP, strconv.Itoa(mapping.rewrittenPort))
		nmd, ok := rpt.Endpoint.NodeMetadatas[realEndpointID]
		if !ok {
			continue
		}

		rpt.Endpoint.NodeMetadatas[copyEndpointID] = nmd.Copy()
	}

	return nil
}

func conntrackModulePresent() bool {
	f, err := os.Open(modules)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, conntrackModule) {
			return true
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("conntrack error: %v", err)
	}

	log.Printf("conntrack: failed to find module %s", conntrackModule)
	return false
}
