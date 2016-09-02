package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
)

// applyTrafficControlRules set the network policies
func applyTrafficControlRules(pid int, rules []string) (netNSID string, err error) {
	cmds := [][]string{
		strings.Fields("tc qdisc replace dev eth0 root handle 1: netem"),
	}
	cmd := strings.Fields("tc qdisc change dev eth0 root handle 1: netem")
	cmd = append(cmd, rules...)
	cmds = append(cmds, cmd)

	netNS := fmt.Sprintf("/proc/%d/ns/net", pid)
	err = ns.WithNetNSPath(netNS, func(hostNS ns.NetNS) error {
		for _, cmd := range cmds {
			if output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
				log.Error(string(output))
				return fmt.Errorf("failed to execute command: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to perform traffic control: %v", err)
	}
	netNSID, err = getNSID(netNS)

	if err != nil {
		return "", err
	}
	return netNSID, nil
}

// ApplyLatency sets the latency
func ApplyLatency(pid int, latency string) error {
	if latency == "" {
		return nil
	}
	rules := strings.Fields(fmt.Sprintf("delay %s", latency))

	// Get cached packet loss
	packetLoss, err := getPacketLoss(pid)
	if err != nil {
		return err
	}
	if packetLoss != "-" {
		rules = append(rules, strings.Fields(fmt.Sprintf("loss %s", packetLoss))...)
	}

	netNSID, err := applyTrafficControlRules(pid, rules)

	// Update cached values
	if trafficControlStatusCache[netNSID] == nil {
		trafficControlStatusCache[netNSID] = TrafficControlStatusInit()
	}
	trafficControlStatusCache[netNSID].SetLatency(latency)
	trafficControlStatusCache[netNSID].SetPacketLoss(packetLoss)

	return nil
}

// ApplyPacketLoss sets the packet loss
func ApplyPacketLoss(pid int, packetLoss string) error {
	if packetLoss == "" {
		return nil
	}
	rules := strings.Fields(fmt.Sprintf("loss %s", packetLoss))

	// Get cached latency
	latency, err := getLatency(pid)
	if err != nil {
		return err
	}
	if latency != "-" {
		rules = append(rules, strings.Fields(fmt.Sprintf("delay %s", latency))...)
	}

	netNSID, err := applyTrafficControlRules(pid, rules)

	// Update cached values
	if trafficControlStatusCache[netNSID] == nil {
		trafficControlStatusCache[netNSID] = TrafficControlStatusInit()
	}
	trafficControlStatusCache[netNSID].SetLatency(latency)
	trafficControlStatusCache[netNSID].SetPacketLoss(packetLoss)

	return nil
}

// ClearTrafficControlSettings clear all parameters of the qdisc with tc
func ClearTrafficControlSettings(pid int) error {
	cmds := [][]string{
		strings.Fields("tc qdisc replace dev eth0 root handle 1: netem"),
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
	// clear cached parameters
	netNSID, err := getNSID(netNS)
	if err != nil {
		log.Error(netNSID)
		return fmt.Errorf("failed to get network namespace ID: %v", err)
	}
	delete(trafficControlStatusCache, netNSID)
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

func getPacketLoss(pid int) (string, error) {
	var status *trafficControlStatus
	var err error
	if status, err = getStatus(pid); err != nil {
		return "-", err
	} else if status == nil {
		return "-", fmt.Errorf("status for PID %d does not exist", pid)
	}
	return status.packetLoss, nil
}

func getStatus(pid int) (*trafficControlStatus, error) {
	netNS := fmt.Sprintf("/proc/%d/ns/net", pid)
	netNSID, err := getNSID(netNS)
	if err != nil {
		log.Error(netNSID)
		return nil, fmt.Errorf("failed to get network namespace ID: %v", err)
	}
	if status, ok := trafficControlStatusCache[netNSID]; ok {
		return status, nil
	}
	cmd := strings.Fields("tc qdisc show dev eth0")
	var output string
	err = ns.WithNetNSPath(netNS, func(hostNS ns.NetNS) error {
		cmdOut, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
		if err != nil {
			log.Error(string(cmdOut))
			output = "-"
			return fmt.Errorf("failed to execute command: tc qdisc show dev eth0: %v", err)
		}
		output = string(cmdOut)
		return nil
	})
	// cache parameters
	trafficControlStatusCache[netNSID] = &trafficControlStatus{
		latency:    parseLatency(output),
		packetLoss: parsePacketLoss(output),
	}
	status, _ := trafficControlStatusCache[netNSID]
	return status, err
}

func parseLatency(statusString string) string {
	return parseAttribute(statusString, "delay")
}

func parsePacketLoss(statusString string) string {
	return parseAttribute(statusString, "loss")
}
func parseAttribute(statusString string, attribute string) string {
	statusStringSplited := strings.Fields(statusString)
	for i, s := range statusStringSplited {
		if s == attribute {
			if i < len(statusStringSplited)-1 {
				return strings.Trim(statusStringSplited[i+1], "\n")
			}
			return "-"
		}
	}
	return "-"
}

func getNSID(nsPath string) (string, error) {
	nsID, err := os.Readlink(nsPath)
	if err != nil {
		log.Error(nsID)
		return "", fmt.Errorf("failed to execute command: tc qdisc show dev eth0: %v", err)
	}
	return nsID[5 : len(nsID)-1], nil
}
