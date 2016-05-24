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

// +build ignore

package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/kinvolk/tcd/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Println("Missing parameter")
		os.Exit(1)
	}

	conn, err := grpc.Dial(os.Args[1],
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			unixAddr, _ := net.ResolveUnixAddr("unix", addr)
			return net.DialUnix("unix", nil, unixAddr)
		}),
		grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c := tcdapi.NewTcdServiceClient(conn)
	defer conn.Close()

	installResp, err := c.InstallMethod(context.Background(), &tcdapi.InstallRequest{
		Container: os.Args[2],
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Printf("installResp: %v\n", installResp)

	ingressResp, err := c.ConfigureIngressMethod(context.Background(), &tcdapi.ConfigureRequest{
		Container: os.Args[2],
		Delay:     30,
		Loss:      0,
		Rate:      800000,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Printf("ingressResp: %v\n", ingressResp)

	egressResp, err := c.ConfigureEgressMethod(context.Background(), &tcdapi.ConfigureRequest{
		Container: os.Args[2],
		Delay:     70,
		Loss:      0,
		Rate:      800000,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	fmt.Printf("egressResp: %v\n", egressResp)

}
