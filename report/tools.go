// Ensure go mod fetches files needed at code generation time;
// the build constraint ensures this file is ignored
//go:build tools
// +build tools

package report

import (
	_ "github.com/ugorji/go/codec/codecgen"
)
