package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
)

// DoTrafficControl is the function that set the parameters of the qdisc with tc
func DoTrafficControl(pid int, latency string, packetLoss string) error {
	if latency == "" && packetLoss == "" {
		// TODO @alepuccetti: return a warning message: "Nothing to do"
		return nil
	}

	var err error
	cmds := [][]string{
		strings.Fields("tc qdisc replace dev eth0 root handle 1: netem"),

		// These steps are not required, since we don't do
		// ingress traffic control, only egress, see the TODO
		// at the beginning of the file.

		//strings.Fields("ip link add ifb0 type ifb"),
		//strings.Fields("ip link set ifb0 up"),
		//strings.Fields("tc qdisc add dev eth0 handle ffff: ingress"),
		//strings.Fields("tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0"),
		//strings.Fields("tc qdisc replace dev ifb0 handle 1:0 root netem"),

		// Add "loss %d%% rate %dkbit" when we add the
		// possibility to control the packet loss and
		// bandwidth. See the TODO at the beginning of the
		// file.

	}
	cmd := strings.Fields("tc qdisc change dev eth0 root handle 1: netem")
	// TODO @alepuccetti: refactor this code
	if latency == "" {
		// packetLoss cannot be empty
		cmd = append(cmd, "loss")
		cmd = append(cmd, packetLoss)
		// get latency from the cache
		if latency, err = getLatency(pid); err != nil {
			return err
		} else if latency != "-" {
			cmd = append(cmd, "delay")
			cmd = append(cmd, latency)
		}
	} else if packetLoss == "" {
		// latency cannot be empty
		cmd = append(cmd, "delay")
		cmd = append(cmd, latency)
		// get packetLoss from the cache
		if packetLoss, err = getPacketLoss(pid); err != nil {
			return err
		} else if packetLoss != "-" {
			cmd = append(cmd, "loss")
			cmd = append(cmd, packetLoss)
		}
	} else {
		// latency and pckLoss are both new
		cmd = append(cmd, "delay")
		cmd = append(cmd, latency)
		cmd = append(cmd, "loss")
		cmd = append(cmd, packetLoss)
	}
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
		return fmt.Errorf("failed to perform traffic control: %v", err)
	}
	// cache parameters
	netNSID, err := getNSID(netNS)
	if err != nil {
		log.Error(netNSID)
		return fmt.Errorf("failed to get network namespace ID: %v", err)
	}
	trafficControlStatusCache[netNSID] = trafficControlStatus{
		latency: func(latency string) string {
			if latency == "" {
				return "-"
			}
			return latency
		}(latency),
		packetLoss: func(packetLoss string) string {
			if packetLoss == "" {
				return "-"
			}
			return packetLoss
		}(packetLoss),
	}
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
		return &emptyTrafficControlStatus, fmt.Errorf("failed to get network namespace ID: %v", err)
	}
	if status, ok := trafficControlStatusCache[netNSID]; ok {
		return &status, nil
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
	trafficControlStatusCache[netNSID] = trafficControlStatus{
		latency:    parseLatency(output),
		packetLoss: parsePacketLoss(output),
	}
	status, _ := trafficControlStatusCache[netNSID]
	return &status, err
}

func parseLatency(statusString string) (string, error) {
	return parseAttribute(statusString, "delay")
}

func parsePacketLoss(statusString string) (string, error) {
	return parseAttribute(statusString, "loss")
}
func parseAttribute(statusString string, attribute string) (string, error) {
	statusStringSplited := strings.Fields(statusString)
	for i, s := range statusStringSplited {
		if s == attribute {
			if i < len(statusStringSplited)-1 {
				return strings.Trim(statusStringSplited[i+1], "\n"), nil
			}
			return "-", nil
		}
	}
	return "-", fmt.Errorf("%s not found", attribute)
}

func getNSID(nsPath string) (string, error) {
	nsID, err := os.Readlink(nsPath)
	if err != nil {
		log.Error(nsID)
		return "", fmt.Errorf("failed to execute command: tc qdisc show dev eth0: %v", err)
	}
	return nsID[5 : len(nsID)-1], nil
}
