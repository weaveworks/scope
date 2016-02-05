package main

import (
	"fmt"
	"os"
	"strings"

	log "github.com/Sirupsen/logrus"
	weavecommon "github.com/weaveworks/weave/common"
)

var version = "dev" // set at build time

type prefixFormatter struct {
	prefix []byte
	next   log.Formatter
}

func (f *prefixFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.next.Format(entry)
	if err != nil {
		return formatted, err
	}
	return append(f.prefix, formatted...), nil
}

func setLogFormatter(prefix string) {
	if !strings.HasSuffix(prefix, " ") {
		prefix += " "
	}
	f := prefixFormatter{
		prefix: []byte(prefix),
		// reuse weave's log format
		next: weavecommon.Log.Formatter,
	}
	log.SetFormatter(&f)
}

func setLogLevel(levelname string) {
	level, err := log.ParseLevel(levelname)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(level)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s (app|probe|version) args...\n", os.Args[0])
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}

	module := os.Args[1]
	os.Args = append([]string{os.Args[0]}, os.Args[2:]...)

	switch module {
	case "app":
		appMain()
	case "probe":
		probeMain()
	case "version":
		fmt.Println("Weave Scope version", version)
	default:
		usage()
	}
}
