package bpf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

/*
#include <sys/types.h>
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
#include <linux/bpf.h>
#include <poll.h>
#include <linux/perf_event.h>
#include <sys/resource.h>

// from https://github.com/safchain/goebpf
// Apache License

// bpf map structure used by C program to define maps and
// used by elf loader.
typedef struct bpf_map_def {
  unsigned int type;
  unsigned int key_size;
  unsigned int value_size;
  unsigned int max_entries;
} bpf_map_def;

typedef struct bpf_map {
	int         fd;
	bpf_map_def def;
} bpf_map;

static __u64 ptr_to_u64(void *ptr)
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

static bpf_map *bpf_load_map(bpf_map_def *map_def)
{
	bpf_map *map;

	map = calloc(1, sizeof(bpf_map));
	if (map == NULL)
		return NULL;

	memcpy(&map->def, map_def, sizeof(bpf_map_def));

	map->fd = bpf_create_map(map_def->type,
		map_def->key_size,
		map_def->value_size,
		map_def->max_entries
	);

	if (map->fd < 0)
		return 0;

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

static int perf_event_open_tracepoint(int tracepoint_id, int pid, int cpu,
                           int group_fd, unsigned long flags)
{
	struct perf_event_attr attr = {0,};
	attr.type = PERF_TYPE_TRACEPOINT;
	attr.sample_type = PERF_SAMPLE_RAW;
	attr.sample_period = 1;
	attr.wakeup_events = 1;
	attr.config = tracepoint_id;

	return syscall(__NR_perf_event_open, &attr, pid, cpu,
                      group_fd, flags);
}

#define PAGE_COUNT 8

typedef int (*consume_fn)(void *data, int size, int callbackIndex);

struct perf_event_sample {
	struct perf_event_header header;
	__u32 size;
	char data[];
};

static int perf_event_read(volatile struct perf_event_mmap_page *header, consume_fn fn, uint64_t callback_data_index)
{
	int consumed = 0;
	int page_size;
	page_size = getpagesize();

	__u64 data_tail = header->data_tail;
	__u64 data_head = header->data_head;
	__u64 buffer_size = PAGE_COUNT * page_size;
	void *base, *begin, *end;
	char buf[256];

	asm volatile("" ::: "memory"); // in real code it should be smp_rmb()
	if (data_head == data_tail)
		return 0;

	base = ((char *)header) + page_size;

	begin = base + data_tail % buffer_size;
	end = base + data_head % buffer_size;

	while (begin != end) {
		struct perf_event_sample *e;

		e = begin;
		if (begin + e->header.size > base + buffer_size) {
			long len = base + buffer_size - begin;

			assert(len < e->header.size);
			memcpy(buf, begin, len);
			memcpy(buf + len, base, e->header.size - len);
			e = (void *) buf;
			begin = base + e->header.size - len;
		} else if (begin + e->header.size == base + buffer_size) {
			begin = base;
		} else {
			begin += e->header.size;
		}

		if (e->header.type == PERF_RECORD_SAMPLE) {
			consumed = fn(e->data, e->size, callback_data_index) || consumed;
		} else if (e->header.type == PERF_RECORD_LOST) {
			struct {
				struct perf_event_header header;
				__u64 id;
				__u64 lost;
			} *lost = (void *) e;
			printf("lost %lld events\n", lost->lost);
		} else {
			printf("unknown event type=%d size=%d\n",
			       e->header.type, e->header.size);
		}
	}

	__sync_synchronize(); // smp_mb()
	header->data_tail = data_head;

	return consumed;
}


extern int callback_to_go(void *, int, uint64_t);
*/
import "C"

type EventCb func([]byte)

var myEventCb EventCb

// BPFMap represents a eBPF map. An eBPF map has to be declared in the C file
type BPFMap struct {
	Name       string
	SectionIdx int
	Idx        int
	m          *C.bpf_map

	// only for perf maps
	pmuFDs  []C.int
	headers []*C.struct_perf_event_mmap_page
}

// BPFKProbe represents a kprobe or kretprobe. they have to be declared in the C file
type BPFKProbe struct {
	Name string
	fd   int
	efd  int
}

type BPFMapIterator struct {
	key interface{}
	m   *BPFMap
}

type BPFKProbePerf struct {
	fileName string
	file     *elf.File

	log    []byte
	maps   map[string]*BPFMap
	probes map[string]*BPFKProbe
}

func NewBpfPerfEvent(fileName string) *BPFKProbePerf {
	return &BPFKProbePerf{
		fileName: fileName,
		maps:     make(map[string]*BPFMap),
		probes:   make(map[string]*BPFKProbe),
		log:      make([]byte, 65536),
	}
}

// from https://github.com/safchain/goebpf
// Apache License

func (b *BPFKProbePerf) readLicense() (string, error) {
	if lsec := b.file.Section("license"); lsec != nil {
		data, err := lsec.Data()
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	return "", nil
}

func (b *BPFKProbePerf) readVersion() (int, error) {
	if vsec := b.file.Section("version"); vsec != nil {
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
		return int(version), nil
	}

	return 0, nil
}

func (b *BPFKProbePerf) readMaps() error {
	for sectionIdx, section := range b.file.Sections {
		if strings.HasPrefix(section.Name, "maps/") {
			data, err := section.Data()
			if err != nil {
				return err
			}

			name := strings.TrimPrefix(section.Name, "maps/")

			mapCount := len(data) / C.sizeof_struct_bpf_map_def
			for i := 0; i < mapCount; i++ {
				pos := i * C.sizeof_struct_bpf_map_def
				cm := C.bpf_load_map((*C.bpf_map_def)(unsafe.Pointer(&data[pos])))
				if cm == nil {
					return fmt.Errorf("Error while loading map %s", section.Name)
				}

				m := &BPFMap{
					Name:       name,
					SectionIdx: sectionIdx,
					Idx:        i,
					m:          cm,
				}

				if oldMap, ok := b.maps[name]; ok {
					return fmt.Errorf("duplicate map: %q (section %q) and %q (section %q)",
						oldMap.Name, b.file.Sections[oldMap.SectionIdx].Name,
						name, section.Name)
				}
				b.maps[name] = m
			}
		}
	}

	return nil
}

func (b *BPFKProbePerf) relocate(data []byte, rdata []byte) error {
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
			return errors.New("Architecture not supported")
		}

		rinsn := (*C.struct_bpf_insn)(unsafe.Pointer(&rdata[offset]))
		if rinsn.code != (C.BPF_LD | C.BPF_IMM | C.BPF_DW) {
			return errors.New("Invalid relocation")
		}

		symbolSec := b.file.Sections[symbol.Section]
		if !strings.HasPrefix(symbolSec.Name, "maps/") {
			fmt.Printf("https://gist.github.com/alban/161ec3c254f05854aeb3ad90730b3fb5\n")
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

func (b *BPFKProbePerf) Load() error {
	fileReader, err := os.Open(b.fileName)
	if err != nil {
		return err
	}

	b.file, err = elf.NewFile(fileReader)
	if err != nil {
		return err
	}

	license, err := b.readLicense()
	if err != nil {
		return err
	}

	lp := unsafe.Pointer(C.CString(license))
	defer C.free(lp)

	version, err := b.readVersion()
	if err != nil {
		return err
	}

	err = b.readMaps()
	if err != nil {
		return err
	}

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

			if isKprobe || isKretprobe {
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

				progFd := C.bpf_prog_load(C.BPF_PROG_TYPE_KPROBE,
					insns, C.int(rsection.Size),
					(*C.char)(lp), C.int(version),
					(*C.char)(unsafe.Pointer(&b.log[0])), C.int(len(b.log)))
				if progFd < 0 {
					return fmt.Errorf("error while loading %q:\n%s", secName, b.log)
				}

				efd, err := b.EnableKprobe(int(progFd), secName, isKretprobe)
				if err != nil {
					return err
				}

				b.probes[secName] = &BPFKProbe{
					Name: secName,
					fd:   int(progFd),
					efd:  efd,
				}
			}
		}
	}

	for i, section := range b.file.Sections {
		if processed[i] {
			continue
		}

		if strings.HasPrefix(section.Name, "kprobe/") || strings.HasPrefix(section.Name, "kretprobe/") {
			data, err := section.Data()
			if err != nil {
				panic(err)
			}

			if len(data) == 0 {
				continue
			}

			insns := (*C.struct_bpf_insn)(unsafe.Pointer(&data[0]))

			fd := C.bpf_prog_load(C.BPF_PROG_TYPE_KPROBE,
				insns, C.int(section.Size),
				(*C.char)(lp), C.int(version),
				(*C.char)(unsafe.Pointer(&b.log[0])), C.int(len(b.log)))
			if fd < 0 {
				return fmt.Errorf("error while loading %q:\n%s", section.Name, b.log)
			}
			b.probes[section.Name] = &BPFKProbe{
				Name: section.Name,
				fd:   int(fd),
			}
		}
	}

	for name, _ := range b.maps {
		var cpu C.int = 0

		for {
			pmuFD := C.perf_event_open_map(-1 /* pid */, cpu /* cpu */, -1 /* group_fd */, C.PERF_FLAG_FD_CLOEXEC)
			if pmuFD < 0 {
				if cpu == 0 {
					return fmt.Errorf("perf_event_open for map error: %v", err)
				}
				break
			}

			// mmap
			pageSize := os.Getpagesize()
			mmapSize := pageSize * (C.PAGE_COUNT + 1)

			base, err := syscall.Mmap(int(pmuFD), 0, mmapSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
			if err != nil {
				return fmt.Errorf("mmap error: %v", err)
			}

			// enable
			_, _, err2 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(pmuFD), C.PERF_EVENT_IOC_ENABLE, 0)
			if err2 != 0 {
				return fmt.Errorf("error enabling perf event: %v", err2)
			}

			// assign perf fd tp map
			ret := C.bpf_update_element(C.int(b.maps[name].m.fd), unsafe.Pointer(&cpu), unsafe.Pointer(&pmuFD), C.BPF_ANY)
			if ret != 0 {
				return fmt.Errorf("cannot assign perf fd to map: %d (cpu %d)", syscall.Errno(ret), cpu)
			}

			b.maps[name].pmuFDs = append(b.maps[name].pmuFDs, pmuFD)
			b.maps[name].headers = append(b.maps[name].headers, (*C.struct_perf_event_mmap_page)(unsafe.Pointer(&base[0])))

			cpu++
		}
	}

	return nil
}

func (b *BPFKProbePerf) EnableKprobe(progFd int, secName string, isKretprobe bool) (int, error) {
	var probeType, funcName string
	if isKretprobe {
		probeType = "r"
		funcName = strings.TrimPrefix(secName, "kretprobe/")
	} else {
		probeType = "p"
		funcName = strings.TrimPrefix(secName, "kprobe/")
	}
	eventName := probeType + funcName

	kprobeEventsFileName := "/sys/kernel/debug/tracing/kprobe_events"
	f, err := os.OpenFile(kprobeEventsFileName, os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		return 0, fmt.Errorf("cannot open kprobe_events: %v\n", err)
	}
	defer f.Close()

	cmd := fmt.Sprintf("%s:%s %s\n", probeType, eventName, funcName)
	_, err = f.WriteString(cmd)
	if err != nil {
		return 0, fmt.Errorf("cannot write %q to kprobe_events: %v\n", cmd, err)
	}

	kprobeIdFile := fmt.Sprintf("/sys/kernel/debug/tracing/events/kprobes/%s/id", eventName)
	kprobeIdBytes, err := ioutil.ReadFile(kprobeIdFile)
	if err != nil {
		return 0, fmt.Errorf("cannot read kprobe id: %v\n", err)
	}
	kprobeId, err := strconv.Atoi(strings.TrimSpace(string(kprobeIdBytes)))
	if err != nil {
		return 0, fmt.Errorf("invalid kprobe id): %v\n", err)
	}

	efd := C.perf_event_open_tracepoint(C.int(kprobeId), -1 /* pid */, 0 /* cpu */, -1 /* group_fd */, C.PERF_FLAG_FD_CLOEXEC)
	if efd < 0 {
		return 0, fmt.Errorf("perf_event_open for kprobe error")
	}

	_, _, err2 := syscall.Syscall(syscall.SYS_IOCTL, uintptr(efd), C.PERF_EVENT_IOC_ENABLE, 0)
	if err2 != 0 {
		return 0, fmt.Errorf("error enabling perf event: %v", err2)
	}

	_, _, err2 = syscall.Syscall(syscall.SYS_IOCTL, uintptr(efd), C.PERF_EVENT_IOC_SET_BPF, uintptr(progFd))
	if err2 != 0 {
		return 0, fmt.Errorf("error enabling perf event: %v", err2)
	}
	return int(efd), nil
}

// Map returns the BPFMap for the given name. The name is the name used for
// the map declaration with the MAP macro is the eBPF C file.
func (b *BPFKProbePerf) Map(name string) *BPFMap {
	return b.maps[name]
}

func perfEventPoll(fds []C.int) error {
	var pfds []C.struct_pollfd

	for i, _ := range fds {
		var pfd C.struct_pollfd

		pfd.fd = fds[i]
		pfd.events = C.POLLIN

		pfds = append(pfds, pfd)
	}
	_, err := C.poll(&pfds[0], C.nfds_t(len(fds)), -1)
	if err != nil {
		return fmt.Errorf("error polling: %v", err.(syscall.Errno))
	}

	return nil
}

// Assume the timestamp is at the beginning of the user struct
type BytesWithTimestamp [][]byte

func (a BytesWithTimestamp) Len() int      { return len(a) }
func (a BytesWithTimestamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BytesWithTimestamp) Less(i, j int) bool {
	return *(*C.uint64_t)(unsafe.Pointer(&a[i][0])) < *(*C.uint64_t)(unsafe.Pointer(&a[j][0]))
}

type callbackData struct {
	b            *BPFKProbePerf
	incoming     BytesWithTimestamp
	receiverChan chan []byte
}

var callbackRegister = make(map[uint64]*callbackData)
var callbackIndex uint64
var mu sync.Mutex

func registerCallback(data *callbackData) uint64 {
	mu.Lock()
	defer mu.Unlock()
	callbackIndex++
	for callbackRegister[callbackIndex] != nil {
		callbackIndex++
	}
	callbackRegister[callbackIndex] = data
	return callbackIndex
}

func unregisterCallback(i uint64) {
	mu.Lock()
	defer mu.Unlock()
	delete(callbackRegister, i)
}

// Gateway function as required with CGO Go >= 1.6
// "If a C-program wants a function pointer, a gateway function has to
// be written. This is because we can't take the address of a Go
// function and give that to C-code since the cgo tool will generate a
// stub in C that should be called."
//export callback_to_go
func callback_to_go(data unsafe.Pointer, size C.int, callbackDataIndex C.uint64_t) C.int {
	callbackData := callbackRegister[uint64(callbackDataIndex)]

	b := C.GoBytes(data, size)

	callbackData.incoming = append(callbackData.incoming, b)

	return 1 // event consumed
}

// nowNanoseconds returns a time that can be compared to bpf_ktime_get_ns()
func nowNanoseconds() uint64 {
	var ts syscall.Timespec
	syscall.Syscall(syscall.SYS_CLOCK_GETTIME, 1 /* CLOCK_MONOTONIC */, uintptr(unsafe.Pointer(&ts)), 0)
	sec, nsec := ts.Unix()
	return 1000*1000*1000*uint64(sec) + uint64(nsec)
}

func (b *BPFKProbePerf) PollStart(mapName string, receiverChan chan []byte) {
	callbackData := &callbackData{
		b:            b,
		receiverChan: receiverChan,
	}
	callbackDataIndex := registerCallback(callbackData)

	if _, ok := b.maps[mapName]; !ok {
		fmt.Fprintf(os.Stderr, "Cannot find map %q. List of found maps:\n", mapName)
		for key, _ := range b.maps {
			fmt.Fprintf(os.Stderr, "%q\n", key)
		}
		os.Exit(1)
	}

	go func() {
		cpuCount := len(b.maps[mapName].pmuFDs)

		for {
			perfEventPoll(b.maps[mapName].pmuFDs)

			for {
				var harvestCount C.int
				beforeHarvest := nowNanoseconds()
				for cpu := 0; cpu < cpuCount; cpu++ {
					harvestCount += C.perf_event_read(b.maps[mapName].headers[cpu],
						(*[0]byte)(C.callback_to_go),
						C.uint64_t(callbackDataIndex))
				}

				sort.Sort(callbackData.incoming)

				for i := 0; i < len(callbackData.incoming); i++ {
					if *(*uint64)(unsafe.Pointer(&callbackData.incoming[0][0])) > beforeHarvest {
						// This record has been sent after the beginning of the harvest. Stop
						// processing here to keep the order. "incoming" is sorted, so the next
						// elements also must not be processed now.
						break
					}
					receiverChan <- callbackData.incoming[0]
					// remove first element
					callbackData.incoming = callbackData.incoming[1:]
				}
				if harvestCount == 0 && len(callbackData.incoming) == 0 {
					break
				}
			}
		}
	}()
}

func (b *BPFKProbePerf) PollStop(mapName string) {
	// TODO
}
