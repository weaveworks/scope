// Copyright 2016 Kinvolk GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcd

import (
	"fmt"
	"net"
	"os"
	"os/exec"

	"github.com/appc/cni/pkg/ns"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/kinvolk/tcd/api"
)

const intro = `
<node>
	<interface name="com.github.kinvolk.tcd">
		<method name="Install">
			<arg name="container" direction="in" type="s"/>
		</method>
		<method name="ConfigureIngress">
			<arg name="container" direction="in" type="s"/>
			<arg name="delay" direction="in" type="u"/>
			<arg name="loss" direction="in" type="u"/>
			<arg name="rate" direction="in" type="u"/>
		</method>
		<method name="ConfigureEgress">
			<arg name="container" direction="in" type="s"/>
			<arg name="delay" direction="in" type="u"/>
			<arg name="loss" direction="in" type="u"/>
			<arg name="rate" direction="in" type="u"/>
		</method>
	</interface>` + introspect.IntrospectDataString + `</node> `

const selfNetNS = "/proc/self/ns/net"

// TCDDBus implements tcdapi.TcdServiceServer interface.
type TCDDBus struct {
	conn  *dbus.Conn
	store *NetNSStore
}

var _ tcdapi.TcdServiceServer = &TCDDBus{}

func (t TCDDBus) InstallMethod(ctx context.Context, request *tcdapi.InstallRequest) (*tcdapi.InstallResponse, error) {
	t.Install(request.Container)
	return &tcdapi.InstallResponse{}, nil
}

func (t TCDDBus) ConfigureIngressMethod(ctx context.Context, request *tcdapi.ConfigureRequest) (*tcdapi.ConfigureResponse, error) {
	t.ConfigureIngress(request.Container, request.Delay, request.Loss, request.Rate)
	return &tcdapi.ConfigureResponse{}, nil
}

func (t TCDDBus) ConfigureEgressMethod(ctx context.Context, request *tcdapi.ConfigureRequest) (*tcdapi.ConfigureResponse, error) {
	t.ConfigureEgress(request.Container, request.Delay, request.Loss, request.Rate)
	return &tcdapi.ConfigureResponse{}, nil
}

func NewTCD(conn *dbus.Conn) (*TCDDBus, error) {
	var err error

	t := new(TCDDBus)
	t.conn = conn

	t.store, err = NewNetNSStore(conn)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Exporting D-Bus service...\n")
	conn.Export(t, "/com/github/kinvolk/tcd", "com.github.kinvolk.tcd")
	conn.Export(introspect.Introspectable(intro), "/com/github/kinvolk/tcd",
		"org.freedesktop.DBus.Introspectable")

	fmt.Printf("Listening on socket...\n")
	l, err := net.Listen("unix", "/run/tcd/tcd.sock")
	if err != nil {
		return nil, err
	}

	fmt.Printf("Creating grpc server...\n")
	publicServer := grpc.NewServer()
	tcdapi.RegisterTcdServiceServer(publicServer, t)

	fmt.Printf("Serving grpc server...\n")
	go publicServer.Serve(l)
	return t, nil
}

func (t TCDDBus) runCmds(container string, cmds []string) *dbus.Error {
	leader, err := t.store.getLeaderFromContainer(container)
	if err != nil {
		fmt.Printf("Error: cannot get container: %q: %v\n", container, err)
		return dbus.NewError("com.github.kinvolk.tcd.PIDNotFound", nil)
	}
	fmt.Printf("container %q has leader %d\n", container, leader)

	err = ns.WithNetNSPath(fmt.Sprintf("/proc/%d/ns/net", leader), true, func(hostNS *os.File) error {
		for _, cmd := range cmds {
			out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
			fmt.Printf("   $ %s\n", cmd)
			fmt.Printf("%s", out)
			if err != nil {
				return dbus.NewError("com.github.kinvolk.tcd.ExecError", nil)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Could not run command: %v\n", err)
		return dbus.NewError("com.github.kinvolk.tcd.ExecError", nil)
	}

	return nil
}

func (t TCDDBus) Install(container string) *dbus.Error {
	cmds := []string{
		"tc qdisc replace dev eth0 root handle 1: netem",
		"ip link add ifb0 type ifb",
		"ip link set ifb0 up",
		"tc qdisc add dev eth0 handle ffff: ingress",
		"tc filter add dev eth0 parent ffff: protocol ip u32 match u32 0 0 action mirred egress redirect dev ifb0",
		"tc qdisc replace dev ifb0 handle 1:0 root netem",
	}
	return t.runCmds(container, cmds)
}

func (t TCDDBus) ConfigureIngress(container string, delay uint32, loss uint32, rate uint32) *dbus.Error {
	cmds := []string{
		fmt.Sprintf("tc qdisc change dev ifb0 handle 1:0 root netem delay %dms loss %d%% rate %dkbit", delay, loss, rate),
	}
	return t.runCmds(container, cmds)
}

func (t TCDDBus) ConfigureEgress(container string, delay uint32, loss uint32, rate uint32) *dbus.Error {
	cmds := []string{
		fmt.Sprintf("tc qdisc change dev eth0 root handle 1: netem delay %dms loss %d%% rate %dkbit", delay, loss, rate),
	}
	return t.runCmds(container, cmds)
}
