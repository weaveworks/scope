package ptrace

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

const (
	ptraceOptions         = syscall.PTRACE_O_TRACESYSGOOD | syscall.PTRACE_O_TRACECLONE
	ptraceTracesysgoodBit = 0x80
)

// PTracer ptrace processed and threads
type PTracer struct {

	// All ptrace calls must come from the
	// same thread.  So we wait on a separate
	// thread.
	ops           chan func()
	stopped       chan stopped
	quit          chan struct{}
	childAttached chan struct{} //used to signal the wait loop

	threads   map[int]*thread
	processes map[int]*process
}

type stopped struct {
	pid    int
	status syscall.WaitStatus
}

// NewPTracer creates a new ptracer.
func NewPTracer() PTracer {
	t := PTracer{
		ops:           make(chan func()),
		stopped:       make(chan stopped),
		quit:          make(chan struct{}),
		childAttached: make(chan struct{}),

		threads:   make(map[int]*thread),
		processes: make(map[int]*process),
	}
	go t.waitLoop()
	go t.loop()
	return t
}

func (t *PTracer) Stop() {
	out := make(chan []int)
	t.ops <- func() {
		pids := []int{}
		for pid := range t.processes {
			pids = append(pids, pid)
		}
		out <- pids
	}
	for _, pid := range <-out {
		t.StopTracing(pid)
	}
	t.quit <- struct{}{}
}

// TraceProcess starts tracing the given pid
func (t *PTracer) TraceProcess(pid int) *process {
	result := make(chan *process)
	t.ops <- func() {
		process := newProcess(pid, t)
		t.processes[pid] = process
		process.trace()
		result <- process
	}
	return <-result
}

// StopTracing stops tracing all threads for the given pid
func (t *PTracer) StopTracing(pid int) error {
	log.Printf("Detaching from %d", pid)

	errors := make(chan error)
	processes := make(chan *process)

	t.ops <- func() {
		// send sigstop to all threads
		process, ok := t.processes[pid]
		if !ok {
			errors <- fmt.Errorf("PID %d not found", pid)
			return
		}

		// This flag tells the thread to detach when it next stops
		process.detaching = true

		// Now send sigstop to all threads.
		for _, thread := range process.threads {
			log.Printf("sending SIGSTOP to %d", thread.tid)
			if err := syscall.Tgkill(pid, thread.tid, syscall.SIGSTOP); err != nil {
				errors <- err
				return
			}
		}

		processes <- process
	}

	select {
	case err := <-errors:
		return err
	case process := <-processes:
		<-process.detached
		return nil
	}
}

// AttachedPIDs list the currently attached processes.
func (t *PTracer) AttachedPIDs() []int {
	result := make(chan []int)
	t.ops <- func() {
		var pids []int
		for pid := range t.processes {
			pids = append(pids, pid)
		}
		result <- pids
	}
	return <-result
}

func (t *PTracer) traceThread(pid int, process *process) *thread {
	result := make(chan *thread)
	t.ops <- func() {
		thread := newThread(pid, process, t)

		err := syscall.PtraceAttach(pid)
		if err != nil {
			log.Printf("Attach %d failed: %v", pid, err)
			return
		}

		var status syscall.WaitStatus
		if _, err = syscall.Wait4(pid, &status, 0, nil); err != nil {
			log.Printf("Wait %d failed: %v", pid, err)
			return
		}

		thread.attached = true

		err = syscall.PtraceSetOptions(pid, ptraceOptions)
		if err != nil {
			log.Printf("SetOptions failed, pid=%d, err=%v", pid, err)
			return
		}

		err = syscall.PtraceSyscall(pid, 0)
		if err != nil {
			log.Printf("PtraceSyscall failed, pid=%d, err=%v", pid, err)
			return
		}

		t.threads[pid] = thread
		result <- thread
		select {
		case t.childAttached <- struct{}{}:
		default:
		}
	}
	return <-result
}

func (t *PTracer) waitLoop() {
	var (
		status syscall.WaitStatus
		pid    int
		err    error
	)

	for {
		log.Printf("Waiting...")
		pid, err = syscall.Wait4(-1, &status, syscall.WALL, nil)
		if err != nil && err.(syscall.Errno) == syscall.ECHILD {
			log.Printf( "No children to wait4")
			<-t.childAttached
			continue
		}

		if err != nil {
			log.Printf(" Wait failed: %v %d", err, err.(syscall.Errno))
			return
		}

		log.Printf(" PID %d stopped with signal %#x", pid, status)
		t.stopped <- stopped{pid, status}
	}
}

func (t *PTracer) loop() {
	runtime.LockOSThread()

	for {
		select {
		case op := <-t.ops:
			op()
		case stopped := <-t.stopped:
			t.handleStopped(stopped.pid, stopped.status)
		case <-t.quit:
			return
		}
	}
}

func (t *PTracer) handleStopped(pid int, status syscall.WaitStatus) {
	signal := syscall.Signal(0)
	target, err := t.thread(pid)
	if err != nil {
		log.Printf("thread failed: %v", err)
		return
	}

	if status.Stopped() && status.StopSignal() == syscall.SIGTRAP|ptraceTracesysgoodBit {
		// pid entered Syscall-enter-stop or syscall-exit-stop
		target.syscallStopped()
	} else if status.Stopped() && status.StopSignal() == syscall.SIGTRAP {
		// pid entered PTRACE_EVENT stop
		switch status.TrapCause() {
		case syscall.PTRACE_EVENT_CLONE:
			err := target.handleClone(pid)
			if err != nil {
				log.Printf("clone failed: %v", err)
				return
			}
		default:
			log.Printf("Unknown PTRACE_EVENT %d for pid %d", status.TrapCause(), pid)
		}
	} else if status.Exited() || status.Signaled() {
		// "tracer can safely assume pid will exit"
		t.threadExited(target)
		return
	} else if status.Stopped() {
		// tracee recieved a non-trace related signal
		signal = status.StopSignal()

		if signal == syscall.SIGSTOP && target.process.detaching {
			t.detachThread(target)
			return
		}
	} else {
		// unknown stop - shouldn't happen!
		log.Printf("Pid %d random stop with status %x", pid, status)
	}

	// Restart stopped caller in syscall trap mode.
	log.Printf("Restarting pid %d with signal %d", pid, int(signal))
	err = syscall.PtraceSyscall(pid, int(signal))
	if err != nil {
		log.Printf("PtraceSyscall failed, pid=%d, err=%v", pid, err)
	}
}

func (t *PTracer) detachThread(thread *thread) {
	syscall.PtraceDetach(thread.tid)
	process := thread.process
	delete(process.threads, thread.tid)
	delete(t.threads, thread.tid)
	if len(process.threads) == 0 {
		delete(t.processes, process.pid)
		close(process.detached)
		log.Printf("Process %d detached", process.pid)
	}
}

func pidForTid(tid int) (pid int, err error) {
	var (
		status  *os.File
		scanner *bufio.Scanner
		splits  []string
	)

	status, err = os.Open(fmt.Sprintf("/proc/%d/status", tid))
	if err != nil {
		return
	}
	defer status.Close()

	scanner = bufio.NewScanner(status)
	for scanner.Scan() {
		splits = strings.Split(scanner.Text(), ":")
		if splits[0] != "Tgid" {
			continue
		}

		pid, err = strconv.Atoi(strings.TrimSpace(splits[1]))
		return
	}

	if err = scanner.Err(); err != nil {
		return
	}

	err = fmt.Errorf("Pid not found for proc %d", tid)
	return
}

func (t *PTracer) thread(tid int) (*thread, error) {
	thread, ok := t.threads[tid]
	if !ok {
		pid, err := pidForTid(tid)
		if err != nil {
			return nil, err
		}

		proc, ok := t.processes[pid]
		if !ok {
			return nil, fmt.Errorf("Got new thread %d for unknown process", tid)
		}

		thread = newThread(tid, proc, t)
		t.threads[tid] = thread
		log.Printf("New thread reported, tid=%d, pid=%d", tid, pid)
	}
	return thread, nil
}

func (t *PTracer) threadExited(thread *thread) {
	thread.handleExit()
	delete(t.threads, thread.tid)
	if thread.process != nil {
		delete(thread.process.threads, thread.tid)
	}
}
