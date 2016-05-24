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

package main

import (
	"fmt"
	"os"

	"github.com/kinvolk/tcd/lib"

	"github.com/godbus/dbus"
)

var conn *dbus.Conn

func main() {
	var err error

	os.MkdirAll("/run/tcd", 0700)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot mkdir /run/tcd: %v\n", err)
		os.Exit(1)
	}

	os.Remove("/run/tcd/tcd.sock")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot remove /run/tcd/tcd.sock: %v\n", err)
		os.Exit(1)
	}

	conn, err = dbus.SystemBus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot connect to system bus: %v\n", err)
		os.Exit(1)
	}

	_, err = tcd.NewTCD(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot start tcd: %v\n", err)
		os.Exit(1)
	}

	// Only request name after TCDDBus has been registered to avoid races
	// from clients
	reply, err := conn.RequestName("com.github.kinvolk.tcd",
		dbus.NameFlagDoNotQueue)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot claim name on system bus: %v\n", err)
	} else if reply != dbus.RequestNameReplyPrimaryOwner {
		fmt.Fprintln(os.Stderr, "name already taken")
	}

	select {}
}
