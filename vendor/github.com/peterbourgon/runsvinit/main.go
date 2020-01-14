package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const etcService = "/etc/service"

func main() {
	log.SetFlags(0)

	runsvdir, err := exec.LookPath("runsvdir")
	if err != nil {
		log.Fatal(err)
	}

	sv, err := exec.LookPath("sv")
	if err != nil {
		log.Fatal(err)
	}

	if fi, err := os.Stat(etcService); err != nil {
		log.Fatal(err)
	} else if !fi.IsDir() {
		log.Fatalf("%s is not a directory", etcService)
	}

	if pid := os.Getpid(); pid != 1 {
		log.Printf("warning: I'm not PID 1, I'm PID %d", pid)
	}

	go reapAll()

	supervisor := cmd(runsvdir, etcService)
	if err := supervisor.Start(); err != nil {
		log.Fatal(err)
	}

	log.Printf("%s started", runsvdir)

	go shutdown(sv, supervisor.Process)

	if err := supervisor.Wait(); err != nil {
		log.Printf("%s exited with error: %v", runsvdir, err)
	} else {
		log.Printf("%s exited cleanly", runsvdir)
	}
}

func reapAll() {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGCHLD)
	for range c {
		go reapOne()
	}
}

// From https://github.com/ramr/go-reaper/blob/master/reaper.go
func reapOne() {
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
		return
	}
	log.Printf("reaped child process %d (%+v)", pid, ws)
}

type signaler interface {
	Signal(os.Signal) error
}

func shutdown(sv string, s signaler) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	sig := <-c
	log.Printf("received %s", sig)

	matches, err := filepath.Glob(filepath.Join(etcService, "*"))
	if err != nil {
		log.Printf("when shutting down services: %v", err)
		return
	}

	var stopped []string
	for _, match := range matches {
		fi, err := os.Stat(match)
		if err != nil {
			log.Printf("%s: %v", match, err)
			continue
		}
		if !fi.IsDir() {
			log.Printf("%s: not a directory", match)
			continue
		}
		service := filepath.Base(match)
		stop := cmd(sv, "stop", service)
		if err := stop.Run(); err != nil {
			log.Printf("%s: %v", strings.Join(stop.Args, " "), err)
			continue
		}
		stopped = append(stopped, service)
	}

	log.Printf("stopped %d: %s", len(stopped), strings.Join(stopped, ", "))
	log.Printf("stopping supervisor with signal %s...", sig)
	if err := s.Signal(sig); err != nil {
		log.Print(err)
	}
	log.Printf("shutdown handler exiting")
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
