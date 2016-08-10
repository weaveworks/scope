package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
)

func DoTrafficControl(pid int, delay uint32) error {
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

		split(fmt.Sprintf("tc qdisc change dev eth0 root handle 1: netem delay %dms", delay)),
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
			latency: delay,
			pakLoss: 0,
		}
	}
	return nil
}

func getLatency(pid int) (string, error) {
	output := "-"
	var err error
	if output, err = getStatus(pid); err != nil {
		return "-", err
	} else if output == "" {
		return "-", fmt.Errorf("Error: output is empty")
	}
	return parseLatency(output)
}

func parseLatency(statusString string) (string, error) {
	statusStringSplited := split(statusString)
	for i, s := range statusStringSplited {
		if s == "delay" {
			if i < len(statusStringSplited)-1 {
				return statusStringSplited[i+1], nil
			} else {
				return "-", nil
			}
		}
	}
	return "N/A", fmt.Errorf("delay not found")
}

func getPktLoss(pid int) (string, error) {
	output := "-"
	var err error
	if output, err = getStatus(pid); err != nil {
		return "-", err
	} else if output == "" {
		return "-", fmt.Errorf("Error: output is empty")
	}
	outputSplited := split(output)
	for i, s := range outputSplited {
		if s == "loss" {
			if i < len(outputSplited)-1 {
				output = outputSplited[i+1]
			} else {
				output = "-"
			}
			return output, nil
		}
	}
	return output, fmt.Errorf("delay not found")
}

func parsePktLoss(statusString string) (string, error) {
	statusStringSplited := split(statusString)
	for i, s := range statusStringSplited {
		if s == "loss" {
			if i < len(statusStringSplited)-1 {
				return statusStringSplited[i+1], nil
			} else {
				return "-", nil
			}
		}
	}
	return "N/A", fmt.Errorf("delay not found")
}

// TODO @alepuccetti: return the ful status structure
func getStatus(pid int) (string, error) {
	netNS := fmt.Sprintf("/proc/%d/ns/net", pid)
	netNSID, err := getNSID(netNS)
	if err != nil {
		log.Error(netNSID)
		return "-", fmt.Errorf("failed to get network namespace ID: %v", err)
	}
	if status, ok := trafficControlStatusCache[netNSID]; ok {
		log.Info("Happy caching")
		return string(status.latency), nil
	}
	log.Info("cache miss: execute tc")
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
	return output, err
}

func getNSID(nsPath string) (string, error) {
	if nsID, err := os.Readlink(nsPath); err != nil {
		log.Error(nsID)
		return "", fmt.Errorf("failed to execute command: tc qdisc show dev eth0: %v", err)
	} else {
		return nsID, nil
	}
}

func split(cmd string) []string {
	return strings.Split(cmd, " ")
}
