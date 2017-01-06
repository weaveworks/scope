package endpoint

import (
	"bufio"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/weaveworks/scope/test"
)

// Obtained though conntrack -E -p tcp -o id and then tweaked
const streamedFlowsSource = `[DESTROY] tcp      6 src=10.0.0.1 dst=127.0.0.1 sport=36826 dport=28106 src=10.0.0.2 dst=127.0.0.2 sport=28107 dport=36826 [ASSURED] id=347275904
    [NEW] tcp      6 120 SYN_SENT src=10.0.0.1 dst=127.0.0.1 sport=36898 dport=28107 [UNREPLIED] src=10.0.0.2 dst=127.0.0.2 sport=28107 dport=36898 id=347275904
 [UPDATE] tcp      6 60 SYN_RECV src=10.0.0.1 dst=127.0.0.1 sport=36898 dport=28107 src=10.0.0.2 dst=127.0.0.2 sport=28107 dport=36898 id=347275904
 [UPDATE] tcp      6 432000 ESTABLISHED src=10.0.0.1 dst=127.0.0.1 sport=36898 dport=28107 src=10.0.0.2 dst=127.0.0.2 sport=28107 dport=36898 [ASSURED] id=347275904`

var wantStreamedFlows = []flow{
	{
		Type: destroyType,
		Original: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.1",
				DstIP: "127.0.0.1",
			},
			Layer4: layer4{
				SrcPort: 36826,
				DstPort: 28106,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.2",
				DstIP: "127.0.0.2",
			},
			Layer4: layer4{
				SrcPort: 28107,
				DstPort: 36826,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID: 347275904,
		},
	},
	{
		Type: newType,
		Original: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.1",
				DstIP: "127.0.0.1",
			},
			Layer4: layer4{
				SrcPort: 36898,
				DstPort: 28107,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.2",
				DstIP: "127.0.0.2",
			},
			Layer4: layer4{
				SrcPort: 28107,
				DstPort: 36898,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID:    347275904,
			State: "SYN_SENT",
		},
	},
	{
		Type: updateType,
		Original: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.1",
				DstIP: "127.0.0.1",
			},
			Layer4: layer4{
				SrcPort: 36898,
				DstPort: 28107,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.2",
				DstIP: "127.0.0.2",
			},
			Layer4: layer4{
				SrcPort: 28107,
				DstPort: 36898,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID:    347275904,
			State: "SYN_RECV",
		},
	},
	{
		Type: updateType,
		Original: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.1",
				DstIP: "127.0.0.1",
			},
			Layer4: layer4{
				SrcPort: 36898,
				DstPort: 28107,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "10.0.0.2",
				DstIP: "127.0.0.2",
			},
			Layer4: layer4{
				SrcPort: 28107,
				DstPort: 36898,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID:    347275904,
			State: "ESTABLISHED",
		},
	},
}

func testFlowDecoding(t *testing.T, source string, want []flow, decoder func(scanner *bufio.Scanner) (flow, error)) {
	scanner := bufio.NewScanner(strings.NewReader(source))
	d := time.Millisecond * 100
	for _, wantFlow := range want {
		haveFlow, err := decoder(scanner)
		if err != nil {
			t.Fatalf("Unexpected decoding error: %v", err)
		}
		test.Poll(t, d, wantFlow, func() interface{} { return haveFlow })
	}

	if _, err := decodeStreamedFlow(scanner); err != io.EOF {
		t.Fatalf("Unexpected error value on empty input: %v", err)
	}
}

func TestStreamedFlowDecoding(t *testing.T) {
	testFlowDecoding(t, streamedFlowsSource, wantStreamedFlows, decodeStreamedFlow)
}

// Obtained through conntrack -L -p tcp -o id
// With SELinux, there is a "secctx="
// After "sudo sysctl net.netfilter.nf_conntrack_acct=1", there is "packets=" and "bytes="
const dumpedFlowsSource = `tcp      6 431998 ESTABLISHED src=10.0.2.2 dst=10.0.2.15 sport=49911 dport=22 src=10.0.2.15 dst=10.0.2.2 sport=22 dport=49911 [ASSURED] mark=0 use=1 id=2993966208
tcp      6 108 ESTABLISHED src=172.17.0.5 dst=172.17.0.2 sport=47010 dport=80 src=172.17.0.2 dst=172.17.0.5 sport=80 dport=47010 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 use=1 id=4001098880
tcp      6 431970 ESTABLISHED src=192.168.35.116 dst=216.58.213.227 sport=49862 dport=443 packets=11 bytes=1337 src=216.58.213.227 dst=192.168.35.116 sport=443 dport=49862 packets=8 bytes=716 [ASSURED] mark=0 secctx=system_u:object_r:unlabeled_t:s0 use=1 id=943643840`

var wantDumpedFlows = []flow{
	{
		Original: meta{
			Layer3: layer3{
				SrcIP: "10.0.2.2",
				DstIP: "10.0.2.15",
			},
			Layer4: layer4{
				SrcPort: 49911,
				DstPort: 22,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "10.0.2.15",
				DstIP: "10.0.2.2",
			},
			Layer4: layer4{
				SrcPort: 22,
				DstPort: 49911,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID:    2993966208,
			State: "ESTABLISHED",
		},
	},
	{
		Original: meta{
			Layer3: layer3{
				SrcIP: "172.17.0.5",
				DstIP: "172.17.0.2",
			},
			Layer4: layer4{
				SrcPort: 47010,
				DstPort: 80,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "172.17.0.2",
				DstIP: "172.17.0.5",
			},
			Layer4: layer4{
				SrcPort: 80,
				DstPort: 47010,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID:    4001098880,
			State: "ESTABLISHED",
		},
	},
	{
		Original: meta{
			Layer3: layer3{
				SrcIP: "192.168.35.116",
				DstIP: "216.58.213.227",
			},
			Layer4: layer4{
				SrcPort: 49862,
				DstPort: 443,
				Proto:   "tcp",
			},
		},
		Reply: meta{
			Layer3: layer3{
				SrcIP: "216.58.213.227",
				DstIP: "192.168.35.116",
			},
			Layer4: layer4{
				SrcPort: 443,
				DstPort: 49862,
				Proto:   "tcp",
			},
		},
		Independent: meta{
			ID:    943643840,
			State: "ESTABLISHED",
		},
	},
}

func TestDumpedFlowDecoding(t *testing.T) {
	testFlowDecoding(t, dumpedFlowsSource, wantDumpedFlows, decodeDumpedFlow)
}
