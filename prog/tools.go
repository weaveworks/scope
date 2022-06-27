// Ensure go mod fetches files needed to build the Docker container;
// the build constraint ensures this file is ignored
//go:build tools
// +build tools

package report

import (
	_ "github.com/mjibson/esc"
	_ "github.com/peterbourgon/runsvinit"
)
