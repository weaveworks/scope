package main

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/containernetworking/cni/pkg/ns"
)

// DoTrafficControl modifies the network interface in a network
// namespace of a given PID to install a given delay.
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
	return nil
}

func split(cmd string) []string {
	return strings.Split(cmd, " ")
}
