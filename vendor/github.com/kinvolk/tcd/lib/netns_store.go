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
	"strconv"
	"strings"

	"github.com/fsouza/go-dockerclient"
	"github.com/godbus/dbus"
)

type NetNSStore struct {
	conn           *dbus.Conn
	machineManager dbus.BusObject
	dockerClient   *docker.Client

	// Use a separate D-Bus connection to receive signals from
	// systemd-machined. See limitations on
	// https://github.com/godbus/dbus/issues/8
	connUpdates           *dbus.Conn
	machineManagerUpdates dbus.BusObject
	ch                    chan *dbus.Signal
}

type machinedMachine struct {
	Name    string
	Class   string
	Service string
	Path    dbus.ObjectPath
	Boo     string `dbus:"-"`
}

func NewNetNSStore(conn *dbus.Conn) (*NetNSStore, error) {
	s := new(NetNSStore)
	s.conn = conn
	s.machineManager = conn.Object("org.freedesktop.machine1", "/org/freedesktop/machine1")

	var listMachines []machinedMachine
	err := s.machineManager.Call("org.freedesktop.machine1.Manager.ListMachines", 0).Store(&listMachines)
	if err != nil {
		return nil, err
	}

	s.dockerClient, err = docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	// TODO: use dbus.SystemBusPrivate() and authenticate
	s.connUpdates, err = dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	s.machineManagerUpdates = s.connUpdates.Object("org.freedesktop.machine1", "/org/freedesktop/machine1")
	s.connUpdates.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',path='/org/freedesktop/machine1',interface='org.freedesktop.machine1.Manager',sender='org.freedesktop.machine1',member='MachineNew'")
	s.connUpdates.BusObject().Call("org.freedesktop.DBus.AddMatch", 0,
		"type='signal',path='/org/freedesktop/machine1',interface='org.freedesktop.machine1.Manager',sender='org.freedesktop.machine1',member='MachineRemoved'")

	s.ch = make(chan *dbus.Signal, 10)
	s.connUpdates.Signal(s.ch)
	go s.serveUpdates()

	//fmt.Printf("list machines: %v\n", listMachines)

	return s, nil
}

func (s NetNSStore) serveUpdates() {
	for {
		signal, ok := <-s.ch
		if !ok {
			return
		}
		// TODO: check sender
		// TODO: keep a list of containers
		if signal.Name == "org.freedesktop.machine1.Manager.MachineNew" {
			fmt.Printf("MachineNew from %s\n", signal.Sender)
		}
		if signal.Name == "org.freedesktop.machine1.Manager.MachineRemoved" {
			fmt.Printf("MachineRemoved\n")
		}
	}
}

func (s NetNSStore) getLeaderFromContainer(container string) (uint32, error) {
	if strings.HasPrefix(container, "pid:") {
		pid, err := strconv.ParseUint(strings.TrimPrefix(container, "pid:"), 10, 32)
		return uint32(pid), err
	}
	if strings.HasPrefix(container, "docker:") {
		dockerContainer, err := s.dockerClient.InspectContainer(strings.TrimPrefix(container, "docker:"))
		if err != nil {
			return 0, err
		}
		return uint32(dockerContainer.State.Pid), nil
	}

	var machineObjPath dbus.ObjectPath

	err := s.machineManager.Call("org.freedesktop.machine1.Manager.GetMachine", 0, container).Store(&machineObjPath)
	if err != nil {
		return 0, err
	}
	machineObj := s.conn.Object("org.freedesktop.machine1", machineObjPath)
	leaderVariant, err := machineObj.GetProperty("org.freedesktop.machine1.Machine.Leader")
	if err != nil {
		return 0, err
	}
	leader := leaderVariant.Value().(uint32)
	return leader, nil
}
