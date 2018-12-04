// Copy a report, decoding and re-encoding it.
package main

import (
	"flag"
	"log"

	"github.com/weaveworks/scope/report"
)

func main() {
	flag.Parse()

	if len(flag.Args()) != 2 {
		log.Fatal("usage: copyreport src.(json|msgpack)[.gz] dst.(json|msgpack)[.gz]")
	}

	rpt, err := report.MakeFromFile(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	if err = rpt.WriteToFile(flag.Arg(1)); err != nil {
		log.Fatal(err)
	}
}
