package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const etcService = "/etc/service"

var (
	debugf = log.Printf
	info   = log.Print
	infof  = log.Printf
	fatal  = log.Fatal
	fatalf = log.Fatalf
)

func main() {
	var (
		reap  = flag.Bool("reap", true, "reap orphan children")
		debug = flag.Bool("debug", false, "log debug information")
	)
	flag.Parse()

	log.SetFlags(0)

	if !*debug {
		debugf = func(string, ...interface{}) {}
	}

	runsvdir, err := exec.LookPath("runsvdir")
	if err != nil {
		fatal(err)
	}

	sv, err := exec.LookPath("sv")
	if err != nil {
		fatal(err)
	}

	if fi, err := os.Stat(etcService); err != nil {
		fatal(err)
	} else if !fi.IsDir() {
		fatalf("%s is not a directory", etcService)
	}

	if pid := os.Getpid(); pid != 1 {
		debugf("warning: I'm not PID 1, I'm PID %d", pid)
	}

	if *reap {
		go reapLoop()
	} else {
		infof("warning: NOT reaping zombies")
	}

	supervisor := cmd(runsvdir, etcService)
	if err := supervisor.Start(); err != nil {
		fatal(err)
	}

	debugf("%s started", runsvdir)

	go shutdown(sv, supervisor.Process)

	if err := supervisor.Wait(); err != nil {
		infof("%s exited with error: %v", runsvdir, err)
	} else {
		debugf("%s exited cleanly", runsvdir)
	}
}

// From https://github.com/ramr/go-reaper/blob/master/reaper.go
func reapLoop() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGCHLD)
	for range c {
		reapChildren()
	}
}

func reapChildren() {
	for {
		var (
			ws  syscall.WaitStatus
			pid int
			err error
		)
		for {
			pid, err = syscall.Wait4(-1, &ws, 0, nil)
			if err != syscall.EINTR {
				break
			}
		}
		if err == syscall.ECHILD {
			return // done
		}
		infof("reaped child process %d (%+v)", pid, ws)
	}
}

type signaler interface {
	Signal(os.Signal) error
}

func shutdown(sv string, s signaler) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	debugf("received %s", sig)

	matches, err := filepath.Glob(filepath.Join(etcService, "*"))
	if err != nil {
		infof("when shutting down services: %v", err)
		return
	}

	var stopped []string
	for _, match := range matches {
		fi, err := os.Stat(match)
		if err != nil {
			infof("%s: %v", match, err)
			continue
		}
		if !fi.IsDir() {
			infof("%s: not a directory", match)
			continue
		}
		service := filepath.Base(match)
		stop := cmd(sv, "stop", service)
		if err := stop.Run(); err != nil {
			infof("%s: %v", strings.Join(stop.Args, " "), err)
			continue
		}
		stopped = append(stopped, service)
	}

	debugf("stopped %d: %s", len(stopped), strings.Join(stopped, ", "))
	debugf("stopping supervisor with signal %s...", sig)
	if err := s.Signal(sig); err != nil {
		info(err)
	}
	debugf("shutdown handler exiting")
}

func cmd(path string, args ...string) *exec.Cmd {
	return &exec.Cmd{
		Path:   path,
		Args:   append([]string{path}, args...),
		Env:    os.Environ(),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
}
