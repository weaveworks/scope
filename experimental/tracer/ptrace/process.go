package ptrace

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"sync"
)

type process struct {
	sync.Mutex
	pid int

	detaching bool
	detached  chan struct{}

	tracer  *PTracer
	threads map[int]*thread
	fds     map[int]*Fd
}

func newProcess(pid int, tracer *PTracer) *process {
	return &process{
		pid:      pid,
		tracer:   tracer,
		threads:  make(map[int]*thread),
		fds:      make(map[int]*Fd),
		detached: make(chan struct{}),
	}
}

func (p *process) trace() {
	go p.loop()
}

// This doesn't actually guarantees we follow all the threads.  Oops.
func (p *process) loop() {
	var (
		attached int
	)
	log.Printf("Tracing process %d", p.pid)

	for {
		ps, err := ioutil.ReadDir(fmt.Sprintf("/proc/%d/task", p.pid))
		if err != nil {
			log.Printf("ReadDir failed, pid=%d, err=%v", p.pid, err)
			return
		}

		attached = 0
		for _, file := range ps {
			pid, err := strconv.Atoi(file.Name())
			if err != nil {
				log.Printf("'%s' is not a pid: %v", file.Name(), err)
				attached++
				continue
			}

			p.Lock()
			t, ok := p.threads[pid]
			if !ok {
				t = p.tracer.traceThread(pid, p)
				p.threads[pid] = t
			}
			p.Unlock()

			if !t.attached {
				continue
			}

			attached++
		}

		// When we successfully attach to all threads
		// we can be sure to catch new clones, so we
		// can quit.
		if attached == len(ps) {
			break
		}
	}

	log.Printf("Successfully attached to %d threads", attached)
}

func (p *process) newThread(thread *thread) {
	p.Lock()
	defer p.Unlock()
	p.threads[thread.tid] = thread
}

func (p *process) newFd(fd *Fd) error {
	_, ok := p.fds[fd.fd]
	if ok {
		return fmt.Errorf("New fd %d, alread exists!", fd.fd)
	}
	p.fds[fd.fd] = fd
	return nil
}
