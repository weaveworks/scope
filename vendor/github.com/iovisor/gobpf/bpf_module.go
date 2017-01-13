// Copyright 2016 PLUMgrid
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bpf

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

/*
#cgo CFLAGS: -I/usr/include/bcc/compat
#cgo LDFLAGS: -lbcc
#include <bcc/bpf_common.h>
#include <bcc/libbpf.h>
void perf_reader_free(void *ptr);
*/
import "C"

// BpfModule type
type BpfModule struct {
	p       unsafe.Pointer
	funcs   map[string]int
	kprobes map[string]unsafe.Pointer
}

type compileRequest struct {
	code   string
	cflags []string
	rspCh  chan *BpfModule
}

var (
	defaultCflags []string
	compileCh     chan compileRequest
	bpfInitOnce   sync.Once
)

func bpfInit() {
	defaultCflags = []string{
		fmt.Sprintf("-DNUMCPUS=%d", runtime.NumCPU()),
	}
	compileCh = make(chan compileRequest)
	go compile()
}

// NewBpfModule constructor
func newBpfModule(code string, cflags []string) *BpfModule {
	cflagsC := make([]*C.char, len(defaultCflags)+len(cflags))
	defer func() {
		for _, cflag := range cflagsC {
			C.free(unsafe.Pointer(cflag))
		}
	}()
	for i, cflag := range cflags {
		cflagsC[i] = C.CString(cflag)
	}
	for i, cflag := range defaultCflags {
		cflagsC[len(cflags)+i] = C.CString(cflag)
	}
	cs := C.CString(code)
	defer C.free(unsafe.Pointer(cs))
	c := C.bpf_module_create_c_from_string(cs, 2, (**C.char)(&cflagsC[0]), C.int(len(cflagsC)))
	if c == nil {
		return nil
	}
	return &BpfModule{
		p:       c,
		funcs:   make(map[string]int),
		kprobes: make(map[string]unsafe.Pointer),
	}
}

func NewBpfModule(code string, cflags []string) *BpfModule {
	bpfInitOnce.Do(bpfInit)
	ch := make(chan *BpfModule)
	compileCh <- compileRequest{code, cflags, ch}
	return <-ch
}

func compile() {
	for {
		req := <-compileCh
		req.rspCh <- newBpfModule(req.code, req.cflags)
	}
}

func (bpf *BpfModule) Close() {
	C.bpf_module_destroy(bpf.p)
	// close the kprobes opened by this module
	for k, v := range bpf.kprobes {
		C.perf_reader_free(v)
		desc := fmt.Sprintf("-:kprobes/%s", k)
		descCS := C.CString(desc)
		C.bpf_detach_kprobe(descCS)
		C.free(unsafe.Pointer(descCS))
	}
	for _, fd := range bpf.funcs {
		syscall.Close(fd)
	}
}

func (bpf *BpfModule) LoadNet(name string) (int, error) {
	return bpf.Load(name, C.BPF_PROG_TYPE_SCHED_ACT)
}
func (bpf *BpfModule) LoadKprobe(name string) (int, error) {
	return bpf.Load(name, C.BPF_PROG_TYPE_KPROBE)
}
func (bpf *BpfModule) Load(name string, progType int) (int, error) {
	fd, ok := bpf.funcs[name]
	if ok {
		return fd, nil
	}
	fd, err := bpf.load(name, progType)
	if err != nil {
		return -1, err
	}
	bpf.funcs[name] = fd
	return fd, nil
}

func (bpf *BpfModule) load(name string, progType int) (int, error) {
	nameCS := C.CString(name)
	defer C.free(unsafe.Pointer(nameCS))
	start := (*C.struct_bpf_insn)(C.bpf_function_start(bpf.p, nameCS))
	size := C.int(C.bpf_function_size(bpf.p, nameCS))
	license := C.bpf_module_license(bpf.p)
	version := C.bpf_module_kern_version(bpf.p)
	if start == nil {
		return -1, fmt.Errorf("BpfModule: unable to find %s", name)
	}
	logbuf := make([]byte, 65536)
	logbufP := (*C.char)(unsafe.Pointer(&logbuf[0]))
	fd := C.bpf_prog_load(uint32(progType), start, size, license, version, logbufP, C.uint(len(logbuf)))
	if fd < 0 {
		msg := string(logbuf[:bytes.IndexByte(logbuf, 0)])
		return -1, fmt.Errorf("Error loading bpf program:\n%s", msg)
	}
	return int(fd), nil
}

var kprobeRegexp = regexp.MustCompile("[+.]")

func (bpf *BpfModule) attachProbe(evName, desc string, fd int) error {
	if _, ok := bpf.kprobes[evName]; ok {
		return nil
	}

	evNameCS := C.CString(evName)
	descCS := C.CString(desc)
	res := C.bpf_attach_kprobe(C.int(fd), evNameCS, descCS, -1, 0, -1, nil, nil)
	C.free(unsafe.Pointer(evNameCS))
	C.free(unsafe.Pointer(descCS))

	if res == nil {
		return fmt.Errorf("Failed to attach BPF kprobe")
	}
	bpf.kprobes[evName] = res
	return nil
}

func (bpf *BpfModule) AttachKprobe(event string, fd int) error {
	evName := "p_" + kprobeRegexp.ReplaceAllString(event, "_")
	desc := fmt.Sprintf("p:kprobes/%s %s", evName, event)

	return bpf.attachProbe(evName, desc, fd)
}

func (bpf *BpfModule) AttachKretprobe(event string, fd int) error {
	evName := "r_" + kprobeRegexp.ReplaceAllString(event, "_")
	desc := fmt.Sprintf("r:kprobes/%s %s", evName, event)

	return bpf.attachProbe(evName, desc, fd)
}

func (bpf *BpfModule) TableSize() uint64 {
	size := C.bpf_num_tables(bpf.p)
	return uint64(size)
}

func (bpf *BpfModule) TableId(name string) C.size_t {
	cs := C.CString(name)
	defer C.free(unsafe.Pointer(cs))
	return C.bpf_table_id(bpf.p, cs)
}

func (bpf *BpfModule) TableDesc(id uint64) map[string]interface{} {
	i := C.size_t(id)
	return map[string]interface{}{
		"name":      C.GoString(C.bpf_table_name(bpf.p, i)),
		"fd":        int(C.bpf_table_fd_id(bpf.p, i)),
		"key_size":  uint64(C.bpf_table_key_size_id(bpf.p, i)),
		"leaf_size": uint64(C.bpf_table_leaf_size_id(bpf.p, i)),
		"key_desc":  C.GoString(C.bpf_table_key_desc_id(bpf.p, i)),
		"leaf_desc": C.GoString(C.bpf_table_leaf_desc_id(bpf.p, i)),
	}
}

func (bpf *BpfModule) TableIter() <-chan map[string]interface{} {
	ch := make(chan map[string]interface{})
	go func() {
		size := C.bpf_num_tables(bpf.p)
		for i := C.size_t(0); i < size; i++ {
			ch <- bpf.TableDesc(uint64(i))
		}
		close(ch)
	}()
	return ch
}
