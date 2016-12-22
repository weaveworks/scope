//+build !linux

package endpoint

import "fmt"

func findBpfObjectFile() (string, error) {
	return "", fmt.Errorf("not supported")
}
