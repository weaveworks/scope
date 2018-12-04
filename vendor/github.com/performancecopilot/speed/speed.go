// Package speed implements a golang client for the Performance Co-Pilot
// instrumentation API.
//
// It is based on the C/Perl/Python API implemented in PCP core as well as the
// Java API implemented by `parfait`, a separate project.
//
// Some examples on using the API are implemented as executable go programs in the
// `examples` subdirectory.
package speed

import (
	"fmt"
	"hash/fnv"
	"os"

	"github.com/pkg/errors"
)

// Version is the last tagged version of the package
const Version = "3.0.0"

var histogramInstances = []string{"min", "max", "mean", "variance", "standard_deviation"}
var histogramIndom *PCPInstanceDomain

// init maintains a central location of all things that happen when the package is initialized
// instead of everything being scattered in multiple source files
func init() {
	if err := initConfig(); err != nil {
		fmt.Fprintln(os.Stderr, errors.Errorf("error initializing config. maybe PCP isn't installed properly"))
	}

	var err error
	histogramIndom, err = NewPCPInstanceDomain("histogram", histogramInstances)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Errorf("could not initialize an instance domain for histograms"))
	}
}

// generate a unique hash for a string of the specified bit length
// NOTE: make sure this is as fast as possible
//
// see: http://programmers.stackexchange.com/a/145633
func hash(s string, b uint32) uint32 {
	h := fnv.New32a()

	_, err := h.Write([]byte(s))
	if err != nil {
		panic(err)
	}

	val := h.Sum32()
	if b == 0 {
		return val
	}

	return val & ((1 << b) - 1)
}
