package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
)

func DoTrafficControl(pid int, latency string) error {
	cmds := [][]string{
		split("tc qdisc replace dev eth0 root handle 1: netem"),

		// These steps are not required, since we don't do
		// ingress traffic control, only egress, see the TODO
		// at the beginning of the file.

		//split("ip link add ifb0 type ifb"),
		//split("ip link set ifb0 up"),
		//split("tc qdisc add dev eth0 handle ffff: ingress"),
		//split("tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0"),
		//split("tc qdisc replace dev ifb0 handle 1:0 root netem"),

		// Add "loss %d%% rate %dkbit" when we add the
		// possibility to control the packet loss and
		// bandwidth. See the TODO at the beginning of the
		// file.

	}
	if latency != "" {
		cmds = append(cmds, split(fmt.Sprintf("tc qdisc change dev eth0 root handle 1: netem delay %s", latency)))
	} else {
		cmds = append(cmds, split("tc qdisc change dev eth0 root handle 1: netem"))
	}

		netNS := fmt.Sprintf("/proc/%d/ns/net", pid)
	err := ns.WithNetNSPath(netNS, func(hostNS ns.NetNS) error {
		for _, cmd := range cmds {
			if output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
				log.Error(string(output))
				return fmt.Errorf("failed to execute command: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to perform traffic control: %v", err)
	}
	// cache parameters
	if netNSID, err := getNSID(netNS); err != nil {
		log.Error(netNSID)
		return fmt.Errorf("failed to get network namespace ID: %v", err)
	} else {
		trafficControlStatusCache[netNSID] = trafficControlStatus{
			latency: func(latency string) string {
				if latency == "" {
					return "-"
				}
				return latency
			}(latency),
			pktLoss: "-",
		}
	}
	return nil
}

func getLatency(pid int) (string, error) {
	var status *trafficControlStatus
	var err error
	if status, err = getStatus(pid); err != nil {
		return "-", err
	} else if status == nil {
		return "-", fmt.Errorf("status for PID %d does not exist", pid)
	}
	return status.latency, nil
}

func getPktLoss(pid int) (string, error) {
	var status *trafficControlStatus
	var err error
	if status, err = getStatus(pid); err != nil {
		return "-", err
	} else if status == nil {
		return "-", fmt.Errorf("status for PID %d does not exist", pid)
	}
	return status.pktLoss, nil
}

func getStatus(pid int) (*trafficControlStatus, error) {
	netNS := fmt.Sprintf("/proc/%d/ns/net", pid)
	netNSID, err := getNSID(netNS)
	if err != nil {
		log.Error(netNSID)
		return &emptyTrafficControlStatus, fmt.Errorf("failed to get network namespace ID: %v", err)
	}
	if status, ok := trafficControlStatusCache[netNSID]; ok {
		return &status, nil
	}
	cmd := split("tc qdisc show dev eth0")
	var output string
	err = ns.WithNetNSPath(netNS, func(hostNS ns.NetNS) error {
		if cmdOut, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			log.Error(string(cmdOut))
			output = "-"
			return fmt.Errorf("failed to execute command: tc qdisc show dev eth0: %v", err)
		} else {
			output = string(cmdOut)
		}
		return nil
	})
	// cache parameters
	if netNSID, err := getNSID(netNS); err != nil {
		log.Error(netNSID)
		return &emptyTrafficControlStatus, fmt.Errorf("failed to get network namespace ID: %v", err)
	} else {
		lat, _ := parseLatency(output)
		pktL, _ := parsePktLoss(output)
		trafficControlStatusCache[netNSID] = trafficControlStatus{
			latency: lat,
			pktLoss: pktL,
		}
	}
	status, _ := trafficControlStatusCache[netNSID]
	return &status, err
}

func parseLatency(statusString string) (string, error) {
	return parseAttribute(statusString, "delay")
}

func parsePktLoss(statusString string) (string, error) {
	return parseAttribute(statusString, "loss")
}
func parseAttribute(statusString string, attribute string) (string, error) {
	statusStringSplited := split(statusString)
	for i, s := range statusStringSplited {
		if s == attribute {
			if i < len(statusStringSplited)-1 {
				return strings.Trim(statusStringSplited[i+1], "\n"), nil
			} else {
				return "-", nil
			}
		}
	}
	return "-", fmt.Errorf("%s not found", attribute)
}

func getNSID(nsPath string) (string, error) {
	if nsID, err := os.Readlink(nsPath); err != nil {
		log.Error(nsID)
		return "", fmt.Errorf("failed to execute command: tc qdisc show dev eth0: %v", err)
	} else {
		return nsID[5:len(nsID)-1], nil
	}
}

func split(cmd string) []string {
	return strings.Split(cmd, " ")
}

func printTrafficControlStatusCache() {
	log.Info(stringTrafficControlStatusCache())
}

func stringTrafficControlStatusCache() string {
	output := ""
	for key, val := range trafficControlStatusCache {
		output = fmt.Sprintf("\n%s %s %s %s \n", output, key, val.latency, val.pktLoss)
	}
	return output
}
