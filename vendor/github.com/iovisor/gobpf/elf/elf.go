// +build linux

// Copyright 2016 Cilium Project
// Copyright 2016 Sylvain Afchain
// Copyright 2016 Kinvolk
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

package elf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/iovisor/gobpf/bpffs"
)

/*
#define _GNU_SOURCE
#include <sys/types.h>
#include <sys/stat.h>
#include <errno.h>
#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <time.h>
#include <assert.h>
#include <sys/socket.h>
#include <linux/unistd.h>
#include "include/bpf.h"
#include <poll.h>
#include <linux/perf_event.h>
#include <sys/resource.h>

// from https://github.com/safchain/goebpf
// Apache License, Version 2.0

#define BUF_SIZE_MAP_NS 256

typedef struct bpf_map_def {
  unsigned int type;
  unsigned int key_size;
  unsigned int value_size;
  unsigned int max_entries;
  unsigned int map_flags;
  unsigned int pinning;
  char namespace[BUF_SIZE_MAP_NS];
} bpf_map_def;

typedef struct bpf_map {
	int         fd;
	bpf_map_def def;
} bpf_map;

extern int bpf_pin_object(int fd, const char *pathname);

__u64 ptr_to_u64(void *ptr)
{
	return (__u64) (unsigned long) ptr;
}

static void bpf_apply_relocation(int fd, struct bpf_insn *insn)
{
	insn->src_reg = BPF_PSEUDO_MAP_FD;
	insn->imm = fd;
}

static int bpf_create_map(enum bpf_map_type map_type, int key_size,
	int value_size, int max_entries)
{
	int ret;
	union bpf_attr attr;
	memset(&attr, 0, sizeof(attr));

	attr.map_type = map_type;
	attr.key_size = key_size;
	attr.value_size = value_size;
	attr.max_entries = max_entries;

	ret = syscall(__NR_bpf, BPF_MAP_CREATE, &attr, sizeof(attr));
	if (ret < 0 && errno == EPERM) {
		// When EPERM is returned, two reasons are possible:
		// 1. user has no permissions for bpf()
		// 2. user has insufficent rlimit for locked memory
		// Unfortunately, there is no api to inspect the current usage of locked
		// mem for the user, so an accurate calculation of how much memory to lock
		// for this new program is difficult to calculate. As a hack, bump the limit
		// to unlimited. If program load fails again, return the error.

		struct rlimit rl = {};
		if (getrlimit(RLIMIT_MEMLOCK, &rl) == 0) {
			rl.rlim_max = RLIM_INFINITY;
			rl.rlim_cur = rl.rlim_max;
			if (setrlimit(RLIMIT_MEMLOCK, &rl) == 0) {
				ret = syscall(__NR_bpf, BPF_MAP_CREATE, &attr, sizeof(attr));
			}
			else {
				printf("setrlimit() failed with errno=%d\n", errno);
				return -1;
			}
		}
	}

	return ret;
}

void create_bpf_obj_get(const char *pathname, void *attr)
{
	union bpf_attr *ptr_bpf_attr;
	ptr_bpf_attr = (union bpf_attr *)attr;
	ptr_bpf_attr->pathname = ptr_to_u64((void *) pathname);
}

int get_pinned_obj_fd(const char *path)
{
	union bpf_attr attr = {};
	create_bpf_obj_get(path, &attr);
	return syscall(__NR_bpf, BPF_OBJ_GET, &attr, sizeof(attr));
}

static bpf_map *bpf_load_map(bpf_map_def *map_def, const char *path)
{
	bpf_map *map;
	struct stat st;
	int ret, do_pin = 0;

	map = calloc(1, sizeof(bpf_map));
	if (map == NULL)
		return NULL;

	memcpy(&map->def, map_def, sizeof(bpf_map_def));

	switch (map_def->pinning) {
	case 1: // PIN_OBJECT_NS
		// TODO to be implemented
		return 0;
	case 2: // PIN_GLOBAL_NS
		if (stat(path, &st) == 0) {
			ret = get_pinned_obj_fd(path);
			if (ret < 0) {
				return 0;
			}
			map->fd = ret;
			return map;
		}
		do_pin = 1;
	}

	map->fd = bpf_create_map(map_def->type,
		map_def->key_size,
		map_def->value_size,
		map_def->max_entries
	);

	if (map->fd < 0) {
		return 0;
	}

	if (do_pin) {
		ret = bpf_pin_object(map->fd, path);
		if (ret < 0) {
			return 0;
		}
	}

	return map;
}

static int bpf_prog_load(enum bpf_prog_type prog_type,
	const struct bpf_insn *insns, int prog_len,
	const char *license, int kern_version,
	char *log_buf, int log_size)
{
	int ret;
	union bpf_attr attr;
	memset(&attr, 0, sizeof(attr));

	attr.prog_type = prog_type;
	attr.insn_cnt = prog_len / sizeof(struct bpf_insn);
	attr.insns = ptr_to_u64((void *) insns);
	attr.license = ptr_to_u64((void *) license);
	attr.log_buf = ptr_to_u64(log_buf);
	attr.log_size = log_size;
	attr.log_level = 1;
	attr.kern_version = kern_version;

	ret = syscall(__NR_bpf, BPF_PROG_LOAD, &attr, sizeof(attr));
	if (ret < 0 && errno == EPERM) {
		// When EPERM is returned, two reasons are possible:
		// 1. user has no permissions for bpf()
		// 2. user has insufficent rlimit for locked memory
		// Unfortunately, there is no api to inspect the current usage of locked
		// mem for the user, so an accurate calculation of how much memory to lock
		// for this new program is difficult to calculate. As a hack, bump the limit
		// to unlimited. If program load fails again, return the error.

		struct rlimit rl = {};
		if (getrlimit(RLIMIT_MEMLOCK, &rl) == 0) {
			rl.rlim_max = RLIM_INFINITY;
			rl.rlim_cur = rl.rlim_max;
			if (setrlimit(RLIMIT_MEMLOCK, &rl) == 0) {
				ret = syscall(__NR_bpf, BPF_PROG_LOAD, &attr, sizeof(attr));
			}
			else {
				printf("setrlimit() failed with errno=%d\n", errno);
				return -1;
			}
		}
	}

	return ret;
}

static int bpf_update_element(int fd, void *key, void *value, unsigned long long flags)
{
	union bpf_attr attr = {
		.map_fd = fd,
		.key = ptr_to_u64(key),
		.value = ptr_to_u64(value),
		.flags = flags,
	};

	return syscall(__NR_bpf, BPF_MAP_UPDATE_ELEM, &attr, sizeof(attr));
}


static int perf_event_open_map(int pid, int cpu, int group_fd, unsigned long flags)
{
	struct perf_event_attr attr = {0,};
	attr.type = PERF_TYPE_SOFTWARE;
	attr.sample_type = PERF_SAMPLE_RAW;
	attr.wakeup_events = 1;

	attr.size = sizeof(struct perf_event_attr);
	attr.config = 10; // PERF_COUNT_SW_BPF_OUTPUT

	return syscall(__NR_perf_event_open, &attr, pid, cpu,
		       group_fd, flags);
}
*/
import "C"

const useCurrentKernelVersion = 0xFFFFFFFE

// Based on https://github.com/safchain/goebpf
// Apache License

func elfReadLicense(file *elf.File) (string, error) {
	if lsec := file.Section("license"); lsec != nil {
		data, err := lsec.Data()
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return "", nil
}

func elfReadVersion(file *elf.File) (uint32, error) {
	if vsec := file.Section("version"); vsec != nil {
		data, err := vsec.Data()
		if err != nil {
			return 0, err
		}
		if len(data) != 4 {
			return 0, errors.New("version is not a __u32")
		}
		version := *(*C.uint32_t)(unsafe.Pointer(&data[0]))
		if err != nil {
			return 0, err
		}
		return uint32(version), nil
	}
	return 0, nil
}

func prepareBPFFS(namespace, name string) (string, error) {
	err := bpffs.Mount()
	if err != nil {
		return "", err
	}
	mapPath := filepath.Join(BPFFSPath, namespace, BPFDirGlobals, name)
	err = os.MkdirAll(filepath.Dir(mapPath), syscall.S_IRWXU)
	if err != nil {
		return "", fmt.Errorf("error creating map directory %q: %v", filepath.Dir(mapPath), err)
	}
	return mapPath, nil
}

func validMapNamespace(namespaceRaw *C.char) (string, error) {
	namespace := C.GoStringN(namespaceRaw, C.int(C.strnlen(namespaceRaw, C.BUF_SIZE_MAP_NS)))
	if namespace == "" || namespace == "." || namespace == ".." {
		return "", fmt.Errorf("namespace must not be %q", namespace)
	}
	if strings.Contains(namespace, "/") {
		return "", fmt.Errorf("no '/' allowed in namespace")
	}
	return namespace, nil
}

func elfReadMaps(file *elf.File) (map[string]*Map, error) {
	maps := make(map[string]*Map)
	for _, section := range file.Sections {
		if !strings.HasPrefix(section.Name, "maps/") {
			continue
		}

		data, err := section.Data()
		if err != nil {
			return nil, err
		}
		if len(data) != C.sizeof_struct_bpf_map_def {
			return nil, fmt.Errorf("only one map with size %d allowed per section", C.sizeof_struct_bpf_map_def)
		}

		name := strings.TrimPrefix(section.Name, "maps/")

		mapDef := (*C.bpf_map_def)(unsafe.Pointer(&data[0]))

		var mapPathC *C.char
		if mapDef.pinning > 0 {
			namespace, err := validMapNamespace(&mapDef.namespace[0])
			if err != nil {
				return nil, err
			}
			mapPath, err := prepareBPFFS(namespace, name)
			if err != nil {
				return nil, fmt.Errorf("error preparing bpf fs: %v", err)
			}
			mapPathC = C.CString(mapPath)
			defer C.free(unsafe.Pointer(mapPathC))
		} else {
			mapPathC = nil
		}

		cm, err := C.bpf_load_map(mapDef, mapPathC)
		if cm == nil {
			return nil, fmt.Errorf("error while loading map %q: %v", section.Name, err)
		}

		if oldMap, ok := maps[name]; ok {
			return nil, fmt.Errorf("duplicate map: %q and %q", oldMap.Name, name)
		}
		maps[name] = &Map{
			Name: name,
			m:    cm,
		}

	}
	return maps, nil
}

func (b *Module) relocate(data []byte, rdata []byte) error {
	var symbol elf.Symbol
	var offset uint64

	symbols, err := b.file.Symbols()
	if err != nil {
		return err
	}

	br := bytes.NewReader(data)

	for {
		switch b.file.Class {
		case elf.ELFCLASS64:
			var rel elf.Rel64
			err := binary.Read(br, b.file.ByteOrder, &rel)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

			symNo := rel.Info >> 32
			symbol = symbols[symNo-1]

			offset = rel.Off
		case elf.ELFCLASS32:
			var rel elf.Rel32
			err := binary.Read(br, b.file.ByteOrder, &rel)
			if err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}

			symNo := rel.Info >> 8
			symbol = symbols[symNo-1]

			offset = uint64(rel.Off)
		default:
			return errors.New("architecture not supported")
		}

		rinsn := (*C.struct_bpf_insn)(unsafe.Pointer(&rdata[offset]))
		if rinsn.code != (C.BPF_LD | C.BPF_IMM | C.BPF_DW) {
			return errors.New("invalid relocation")
		}

		symbolSec := b.file.Sections[symbol.Section]
		if !strings.HasPrefix(symbolSec.Name, "maps/") {
			return fmt.Errorf("map location not supported: map %q is in section %q instead of \"maps/%s\"",
				symbol.Name, symbolSec.Name, symbol.Name)
		}
		name := strings.TrimPrefix(symbolSec.Name, "maps/")

		m := b.Map(name)
		if m == nil {
			return fmt.Errorf("relocation error, symbol %q not found in section %q",
				symbol.Name, symbolSec.Name)
		}

		C.bpf_apply_relocation(m.m.fd, rinsn)
	}
}

type SectionParams struct {
	PerfRingBufferPageCount   int
	SkipPerfMapInitialization bool
}

// Load loads the BPF programs and BPF maps in the module. Each ELF section
// can optionally have parameters that changes how it is configured.
func (b *Module) Load(parameters map[string]SectionParams) error {
	if b.fileName != "" {
		fileReader, err := os.Open(b.fileName)
		if err != nil {
			return err
		}
		defer fileReader.Close()
		b.fileReader = fileReader
	}

	var err error
	b.file, err = elf.NewFile(b.fileReader)
	if err != nil {
		return err
	}

	license, err := elfReadLicense(b.file)
	if err != nil {
		return err
	}

	lp := unsafe.Pointer(C.CString(license))
	defer C.free(lp)

	version, err := elfReadVersion(b.file)
	if err != nil {
		return err
	}
	if version == useCurrentKernelVersion {
		version, err = currentVersion()
		if err != nil {
			return err
		}
	}

	maps, err := elfReadMaps(b.file)
	if err != nil {
		return err
	}
	b.maps = maps

	processed := make([]bool, len(b.file.Sections))
	for i, section := range b.file.Sections {
		if processed[i] {
			continue
		}

		data, err := section.Data()
		if err != nil {
			return err
		}

		if len(data) == 0 {
			continue
		}

		if section.Type == elf.SHT_REL {
			rsection := b.file.Sections[section.Info]

			processed[i] = true
			processed[section.Info] = true

			secName := rsection.Name

			isKprobe := strings.HasPrefix(secName, "kprobe/")
			isKretprobe := strings.HasPrefix(secName, "kretprobe/")
			isCgroupSkb := strings.HasPrefix(secName, "cgroup/skb")
			isCgroupSock := strings.HasPrefix(secName, "cgroup/sock")

			var progType uint32
			switch {
			case isKprobe:
				fallthrough
			case isKretprobe:
				progType = uint32(C.BPF_PROG_TYPE_KPROBE)
			case isCgroupSkb:
				progType = uint32(C.BPF_PROG_TYPE_CGROUP_SKB)
			case isCgroupSock:
				progType = uint32(C.BPF_PROG_TYPE_CGROUP_SOCK)
			}

			if isKprobe || isKretprobe || isCgroupSkb || isCgroupSock {
				rdata, err := rsection.Data()
				if err != nil {
					return err
				}

				if len(rdata) == 0 {
					continue
				}

				err = b.relocate(data, rdata)
				if err != nil {
					return err
				}

				insns := (*C.struct_bpf_insn)(unsafe.Pointer(&rdata[0]))

				progFd := C.bpf_prog_load(progType,
					insns, C.int(rsection.Size),
					(*C.char)(lp), C.int(version),
					(*C.char)(unsafe.Pointer(&b.log[0])), C.int(len(b.log)))
				if progFd < 0 {
					return fmt.Errorf("error while loading %q:\n%s", secName, b.log)
				}

				switch {
				case isKprobe:
					fallthrough
				case isKretprobe:
					b.probes[secName] = &Kprobe{
						Name:  secName,
						insns: insns,
						fd:    int(progFd),
					}
				case isCgroupSkb:
					fallthrough
				case isCgroupSock:
					b.cgroupPrograms[secName] = &CgroupProgram{
						Name:  secName,
						insns: insns,
						fd:    int(progFd),
					}
				}
			}
		}
	}

	for i, section := range b.file.Sections {
		if processed[i] {
			continue
		}

		secName := section.Name

		isKprobe := strings.HasPrefix(secName, "kprobe/")
		isKretprobe := strings.HasPrefix(secName, "kretprobe/")
		isCgroupSkb := strings.HasPrefix(secName, "cgroup/skb")
		isCgroupSock := strings.HasPrefix(secName, "cgroup/sock")

		var progType uint32
		switch {
		case isKprobe:
			fallthrough
		case isKretprobe:
			progType = uint32(C.BPF_PROG_TYPE_KPROBE)
		case isCgroupSkb:
			progType = uint32(C.BPF_PROG_TYPE_CGROUP_SKB)
		case isCgroupSock:
			progType = uint32(C.BPF_PROG_TYPE_CGROUP_SOCK)
		}

		if isKprobe || isKretprobe || isCgroupSkb || isCgroupSock {
			data, err := section.Data()
			if err != nil {
				return err
			}

			if len(data) == 0 {
				continue
			}

			insns := (*C.struct_bpf_insn)(unsafe.Pointer(&data[0]))

			progFd := C.bpf_prog_load(progType,
				insns, C.int(section.Size),
				(*C.char)(lp), C.int(version),
				(*C.char)(unsafe.Pointer(&b.log[0])), C.int(len(b.log)))
			if progFd < 0 {
				return fmt.Errorf("error while loading %q:\n%s", section.Name, b.log)
			}

			switch {
			case isKprobe:
				fallthrough
			case isKretprobe:
				b.probes[secName] = &Kprobe{
					Name:  secName,
					insns: insns,
					fd:    int(progFd),
				}
			case isCgroupSkb:
				fallthrough
			case isCgroupSock:
				b.cgroupPrograms[secName] = &CgroupProgram{
					Name:  secName,
					insns: insns,
					fd:    int(progFd),
				}
			}
		}
	}

	return b.initializePerfMaps(parameters)
}

func (b *Module) initializePerfMaps(parameters map[string]SectionParams) error {
	for name, m := range b.maps {
		var cpu C.int = 0

		if m.m != nil && m.m.def._type != C.BPF_MAP_TYPE_PERF_EVENT_ARRAY {
			continue
		}

		pageSize := os.Getpagesize()
		b.maps[name].pageCount = 8 // reasonable default

		sectionName := "maps/" + name
		if params, ok := parameters[sectionName]; ok {
			if params.SkipPerfMapInitialization {
				continue
			}
			if params.PerfRingBufferPageCount > 0 {
				if params.PerfRingBufferPageCount <= 0 || (params.PerfRingBufferPageCount&(params.PerfRingBufferPageCount-1)) != 0 {
					return fmt.Errorf("number of pages (%d) must be stricly positive and a power of 2", params.PerfRingBufferPageCount)
				}
				b.maps[name].pageCount = params.PerfRingBufferPageCount
			}
		}

		for {
			pmuFD, err := C.perf_event_open_map(-1 /* pid */, cpu /* cpu */, -1 /* group_fd */, C.PERF_FLAG_FD_CLOEXEC)
			if pmuFD < 0 {
				if cpu == 0 {
					return fmt.Errorf("perf_event_open for map error: %v", err)
				}
				break
			}

			// mmap
			mmapSize := pageSize * (b.maps[name].pageCount + 1)

			base, err := syscall.Mmap(int(pmuFD), 0, mmapSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
			if err != nil {
				return fmt.Errorf("mmap error: %v", err)
			}

			// enable
			_, _, err2 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(pmuFD), C.PERF_EVENT_IOC_ENABLE, 0)
			if err2 != 0 {
				return fmt.Errorf("error enabling perf event: %v", err2)
			}

			// assign perf fd to map
			ret, err := C.bpf_update_element(C.int(b.maps[name].m.fd), unsafe.Pointer(&cpu), unsafe.Pointer(&pmuFD), C.BPF_ANY)
			if ret != 0 {
				return fmt.Errorf("cannot assign perf fd to map %q: %v (cpu %d)", name, err, cpu)
			}

			b.maps[name].pmuFDs = append(b.maps[name].pmuFDs, pmuFD)
			b.maps[name].headers = append(b.maps[name].headers, (*C.struct_perf_event_mmap_page)(unsafe.Pointer(&base[0])))

			cpu++
		}
	}

	return nil
}

// Map represents a eBPF map. An eBPF map has to be declared in the
// C file.
type Map struct {
	Name string
	m    *C.bpf_map

	// only for perf maps
	pmuFDs    []C.int
	headers   []*C.struct_perf_event_mmap_page
	pageCount int
}

func (b *Module) IterMaps() <-chan *Map {
	ch := make(chan *Map)
	go func() {
		for name := range b.maps {
			ch <- b.maps[name]
		}
		close(ch)
	}()
	return ch
}

func (b *Module) Map(name string) *Map {
	return b.maps[name]
}

func (m *Map) Fd() int {
	return int(m.m.fd)
}
