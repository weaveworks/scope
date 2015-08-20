package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bluele/gcache"
)

const cacheSize = 500

// On every connect and accept, we lookup the local addr
// As this is expensive, we cache the result
var fdAddrCache = gcache.New(cacheSize).LRU().Expiration(15 * time.Second).Build()

type fdCacheKey struct {
	pid int
	fd  int
}
type fdCacheValue struct {
	addr uint32
	port uint16
}

func getCachedLocalAddr(pid, fd int) (uint32, uint16, error) {
	key := fdCacheKey{pid, fd}
	val, err := fdAddrCache.Get(key)
	if val != nil {
		return val.(fdCacheValue).addr, val.(fdCacheValue).port, nil
	}

	addr, port, err := getLocalAddr(pid, fd)
	if err != nil {
		return 0, 0, err
	}
	fdAddrCache.Set(key, fdCacheValue{addr, port})
	return addr, port, nil
}

// On every connect or accept, we cache the syscall that caused
// it for matching with a connection from conntrack
var syscallCache = gcache.New(cacheSize).LRU().Expiration(15 * time.Second).Build()

type syscallCacheKey struct {
	localAddr uint32
	localPort uint16
}
type syscallCacheValue *syscall

// One ever conntrack connection, we cache it by local addr, port to match with
// a future syscall
var conntrackCache = gcache.New(cacheSize).LRU().Expiration(15 * time.Second).Build()

type conntrackCacheKey syscallCacheKey

// And keep a list of successfully matched connection, for us to close out
// when we get the close syscall

func main() {
	ftrace, err := NewFtracer()
	if err != nil {
		panic(err)
	}

	ftrace.start()
	defer ftrace.stop()

	syscalls := make(chan *syscall, 100)
	go ftrace.events(syscalls)

	onConnection := func(s *syscall) {
		fdStr, ok := s.args["fd"]
		if !ok {
			panic("no pid")
		}
		fd64, err := strconv.ParseInt(fdStr, 32, 16)
		if err != nil {
			panic(err)
		}
		fd := int(fd64)

		addr, port, err := getCachedLocalAddr(s.pid, fd)
		if err != nil {
			fmt.Printf("Failed to get local addr for pid=%d fd=%d: %v\n", s.pid, fd, err)
			return
		}

		fmt.Printf("%+v %d %d\n", s, addr, port)
		syscallCache.Set(syscallCacheKey{addr, port}, s)
	}

	onAccept := func(s *syscall) {

	}

	onClose := func(s *syscall) {

	}

	fmt.Println("Started")

	for {
		select {
		case s := <-syscalls:

			switch s.name {
			case "connect":
				onConnection(s)
			case "accept", "accept4":
				onAccept(s)
			case "close":
				onClose(s)
			}
		}
	}
}
