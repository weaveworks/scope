package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
)

const (
	perms = 0777
)

var (
	lineMatcher  = regexp.MustCompile(`^\s*[a-z\-]+\-(\d+)\s+\[(\d{3})] (?:\.|1){4} ([\d\.]+): (.*)$`)
	enterMatcher = regexp.MustCompile(`^([\w_]+)\((.*)\)$`)
	argMatcher   = regexp.MustCompile(`(\w+): (\w+)`)
	exitMatcher  = regexp.MustCompile(`^([\w_]+) -> (\w+)$`)
)

// Ftrace is a tracer using ftrace...
type Ftrace struct {
	ftraceRoot  string
	root        string
	outstanding map[int]*syscall // map from pid (readlly tid) to outstanding syscall
}

type event struct {
	pid int
	cpu int
	ts  float64
}

type syscall struct {
	event
	name       string
	args       map[string]string
	returnCode int64
}

func findDebugFS() (string, error) {
	contents, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		return "", err
	}
	scanner := bufio.NewScanner(bytes.NewBuffer(contents))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		if fields[2] == "debugfs" {
			return fields[1], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("Not found")
}

// NewFtrace constucts a new Ftrace instance.
func NewFtrace() (*Ftrace, error) {
	root, err := findDebugFS()
	if err != nil {
		return nil, err
	}
	scopeRoot := path.Join(root, "tracing", "instances", "scope")
	if err := os.Mkdir(scopeRoot, perms); err != nil && os.IsExist(err) {
		if err := os.Remove(scopeRoot); err != nil {
			return nil, err
		}
		if err := os.Mkdir(scopeRoot, perms); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	return &Ftrace{
		ftraceRoot:  root,
		root:        scopeRoot,
		outstanding: map[int]*syscall{},
	}, nil
}

func (f *Ftrace) destroy() error {
	return os.Remove(f.root)
}

func (f *Ftrace) enableTracing() error {
	// need to enable tracing at root to get trace_pipe to block in my instance. Weird
	if err := ioutil.WriteFile(path.Join(f.ftraceRoot, "tracing", "tracing_on"), []byte("1"), perms); err != nil {
		return err
	}
	return ioutil.WriteFile(path.Join(f.root, "tracing_on"), []byte("1"), perms)
}

func (f *Ftrace) disableTracing() error {
	if err := ioutil.WriteFile(path.Join(f.root, "tracing_on"), []byte("0"), perms); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(f.ftraceRoot, "tracing", "tracing_on"), []byte("1"), perms)
}

func (f *Ftrace) enableEvent(class, event string) error {
	return ioutil.WriteFile(path.Join(f.root, "events", class, event, "enable"), []byte("1"), perms)
}

func mustAtoi(a string) int {
	i, err := strconv.Atoi(a)
	if err != nil {
		panic(err)
	}
	return i
}

func mustAtof(a string) float64 {
	i, err := strconv.ParseFloat(a, 64)
	if err != nil {
		panic(err)
	}
	return i
}

func (f *Ftrace) events() error {
	file, err := os.Open(path.Join(f.root, "trace_pipe"))
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := lineMatcher.FindStringSubmatch(scanner.Text())
		if matches == nil {
			continue
		}
		e := event{
			pid: mustAtoi(matches[1]),
			cpu: mustAtoi(matches[2]),
			ts:  mustAtof(matches[3]),
		}
		f.handleEvent(e, matches[4])
	}

	return scanner.Err()
}

func (f *Ftrace) handleEvent(e event, log string) {
	if matches := enterMatcher.FindStringSubmatch(log); matches != nil {
		name := matches[1]
		args := map[string]string{}
		for _, arg := range argMatcher.FindAllStringSubmatch(matches[2], -1) {
			args[arg[1]] = arg[2]
		}

		s := &syscall{
			event: e,
			name:  name,
			args:  args,
		}

		f.outstanding[e.pid] = s
		return
	}

	if matches := exitMatcher.FindStringSubmatch(log); matches != nil {
		s, ok := f.outstanding[e.pid]
		if !ok {
			return
		}
		delete(f.outstanding, e.pid)
		returnCode, err := strconv.ParseUint(matches[2], 0, 64)
		if err != nil {
			panic(err)
		}
		s.returnCode = int64(returnCode)
		fmt.Printf("%+v\n", *s)
		return
	}
}

func main() {
	ftrace, err := NewFtrace()
	if err != nil {
		panic(err)
	}
	defer ftrace.destroy()

	for _, e := range []struct{ class, event string }{
		{"syscalls", "sys_enter_connect"},
		{"syscalls", "sys_exit_connect"},
		{"syscalls", "sys_enter_accept"},
		{"syscalls", "sys_exit_accept"},
		{"syscalls", "sys_enter_accept4"},
		{"syscalls", "sys_exit_accept4"},
	} {
		if err := ftrace.enableEvent(e.class, e.event); err != nil {
			panic(err)
		}
	}

	if err := ftrace.enableTracing(); err != nil {
		panic(err)
	}
	if err := ftrace.events(); err != nil {
		panic(err)
	}
	fmt.Println("done")
	defer ftrace.disableTracing()
}
