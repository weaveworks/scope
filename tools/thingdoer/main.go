package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"gopkg.in/yaml.v2"
)

type thing struct {
	Name    string   `yaml:"name"`
	After   []string `yaml:"after"`
	Command string
	success bool
}

func (t *thing) run() {
	var out bytes.Buffer

	cmd := exec.Command("/bin/bash", "-c", t.Command)
	cmd.Env = os.Environ()
	cmd.Stdout = &out
	cmd.Stderr = &out

	start := time.Now()
	err := cmd.Run()
	duration := float64(time.Now().Sub(start)) / float64(time.Second)

	if err != nil {
		log.Printf(">>> %s finished after %0.1f secs with error: %v\n", t.Name, duration, err)
	} else {
		log.Printf(">>> %s finished with success after %0.1f secs\n", t.Name, duration)
	}

	log.Print(out.String())
	log.Println()

	t.success = err == nil
}

func remove(s string, ts []string) []string {
	for i, t := range ts {
		if t == s {
			return ts[:i+copy(ts[i:], ts[i+1:])]
		}
	}
	return ts
}

func main() {
	var parallelism int
	flag.IntVar(&parallelism, "p", 2, "level of parallelism")
	flag.Parse()

	args := flag.Args()
	if len(args) <= 0 {
		log.Fatalf("Usage: thingdoer <file>")
	}

	contents, err := ioutil.ReadFile(args[0])
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	things := []*thing{}
	if err := yaml.Unmarshal(contents, &things); err != nil {
		log.Fatalf("Error parsing file: %v", err)
	}

	log.Printf("things: %#v", things)

	next := make(chan *thing, parallelism)
	done := make(chan *thing)
	wg := sync.WaitGroup{}
	wg.Add(parallelism)

	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()

			for thing := range next {
				thing.run()
				done <- thing
			}
		}()
	}

	sched := func() {
		var thing *thing
		for i, t := range things {
			if len(t.After) == 0 {
				thing = t
				things = things[:i+copy(things[i:], things[i+1:])]
				break
			}
		}
		if thing != nil {
			log.Printf("Doing %s", thing.Name)
			next <- thing
		} else {
			log.Printf("Nothing to do, waiting for something to finish")
		}
	}

	// schedule p things if possible
	for i := 0; i < parallelism; i++ {
		sched()
	}

	for {
		// next, wait for something to finish
		thing := <-done
		if !thing.success {
			log.Printf("%s failed, stopping", thing.Name)
			break
		}

		// remove it as a dependency from
		// other things
		for _, t := range things {
			t.After = remove(thing.Name, t.After)
		}

		// (potentially) schedule something else to run
		if len(things) > 0 {
			sched()
		} else {
			break
		}
	}

	close(next)
	wg.Wait()
}
