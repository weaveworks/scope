package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/performancecopilot/speed/mmvdump"
)

func main() {
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Println("usage: mmvdump <file>")
		return
	}

	file := flag.Arg(0)
	d, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	header, tocs, metrics, values, instances, indoms, strings, err := mmvdump.Dump(d)
	if err != nil {
		panic(err)
	}

	fmt.Printf("File      = %v\n", file)
	if err := mmvdump.Write(os.Stdout, header, tocs, metrics, values, instances, indoms, strings); err != nil {
		panic(err)
	}
}
