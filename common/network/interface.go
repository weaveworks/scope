package network

import (
	"fmt"
	"net"
)

// GetFirstAddressOf returns the first address of the supplied interface name.
func GetFirstAddressOf(name string) (string, error) {
	inf, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}

	addrs, err := inf.Addrs()
	if err != nil {
		return "", err
	}
	if len(addrs) <= 0 {
		return "", fmt.Errorf("No address found for %s", name)
	}

	switch v := addrs[0].(type) {
	case *net.IPNet:
		return v.IP.String(), nil
	default:
		return "", fmt.Errorf("No address found for %s", name)
	}
}
