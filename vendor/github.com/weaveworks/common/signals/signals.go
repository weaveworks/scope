package signals

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// SignalReceiver represents a subsystem/server/... that can be stopped or
// queried about the status with a signal
type SignalReceiver interface {
	Stop() error
}

// Logger is something to log too.
type Logger interface {
	Infof(format string, args ...interface{})
}

// SignalHandlerLoop blocks until it receives a SIGINT, SIGTERM or SIGQUIT.
// For SIGINT and SIGTERM, it exits; for SIGQUIT is print a goroutine stack
// dump.
func SignalHandlerLoop(log Logger, ss ...SignalReceiver) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	buf := make([]byte, 1<<20)
	for {
		switch <-sigs {
		case syscall.SIGINT, syscall.SIGTERM:
			log.Infof("=== received SIGINT/SIGTERM ===\n*** exiting")
			for _, subsystem := range ss {
				subsystem.Stop()
			}
			return
		case syscall.SIGQUIT:
			stacklen := runtime.Stack(buf, true)
			log.Infof("=== received SIGQUIT ===\n*** goroutine dump...\n%s\n*** end", buf[:stacklen])
		}
	}
}
