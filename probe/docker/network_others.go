// +build !linux

package docker

import (
	"errors"
	"net"
)

func namespaceIPAddresses(processID int) ([]*net.IPNet, error) {
	return nil, errors.New("namespaceIPAddresses not implemented on this platform")
}
