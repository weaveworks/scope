package cri

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"google.golang.org/grpc"

	client "github.com/weaveworks/scope/cri/runtime"
)

const unixProtocol = "unix"

func dial(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(unixProtocol, addr, timeout)
}

func getAddressAndDialer(endpoint string) (string, func(addr string, timeout time.Duration) (net.Conn, error), error) {
	addr, err := parseEndpointWithFallbackProtocol(endpoint, unixProtocol)
	if err != nil {
		return "", nil, err
	}

	return addr, dial, nil
}

func parseEndpointWithFallbackProtocol(endpoint string, fallbackProtocol string) (addr string, err error) {
	var protocol string

	protocol, addr, err = parseEndpoint(endpoint)

	if err != nil {
		return "", err
	}

	if protocol == "" {
		fallbackEndpoint := fallbackProtocol + "://" + endpoint
		_, addr, err = parseEndpoint(fallbackEndpoint)

		if err != nil {
			return "", err
		}
	}
	return addr, err
}

func parseEndpoint(endpoint string) (string, string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", "", err
	}

	if u.Scheme == "tcp" {
		return "tcp", u.Host, fmt.Errorf("endpoint was not unix socket %v", u.Scheme)
	} else if u.Scheme == "unix" {
		return "unix", u.Path, nil
	} else if u.Scheme == "" {
		return "", "", nil
	} else {
		return u.Scheme, "", fmt.Errorf("protocol %q not supported", u.Scheme)
	}
}

// NewCRIClient creates client to CRI.
func NewCRIClient(endpoint string) (client.RuntimeServiceClient, error) {
	addr, dailer, err := getAddressAndDialer(endpoint)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithDialer(dailer))
	if err != nil {
		return nil, err
	}

	return client.NewRuntimeServiceClient(conn), nil
}
