// Ensure go mod fetches files needed to build the Docker container;
// the build constraint ensures this file is ignored
// +build tools

package report

import (
	_ "github.com/peterbourgon/runsvinit"
)
