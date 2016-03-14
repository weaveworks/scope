// Package ionet is a bridge between the stdlib's net and io packages.
//
// ionet provides a net.Conn and a net.Listener in which connections
// use an io.Reader and an io.Writer instead of a traditional network stack.
//
// This can be handy in unit tests, because it enables you to mock out
// the network.
//
// It's also useful when using an external network stack. At PayPal, ionet
// is used in PayPal Beacon. Beacon uses a Bluetooth Low Energy chip accessed
// over a serial connection. ionet enables the use of net-based code, such as
// the stdlib's net/http, with a mediated network.
package ionet

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"
)

// Conn is a net.Conn backed by an io.Reader and an io.Writer.
// The zero value for Conn uses a reader that always returns EOF
// and ioutil.Discard as a writer.
//
// "Reader" and "Writer" are relative to which half of
// the connection you are on. R and W in Conn are named from
// the server's (listener's) perspective. That is, the server reads from R
// and writes to W; the client (dialer) does the opposite. See also the
// documentation for Dial.
//
// Conn serializes reads and writes to R and W, so R and W do
// not need to be concurrency-safe. After being closed, no
// new reads/writes will be issued to R or W. However, reads/writes
// that were requested before closing (and which were perhaps blocked)
// may still be passed to R or W.
type Conn struct {
	R   io.Reader
	W   io.Writer
	rmu sync.Mutex // used to serialize reads from R (io.Reader is not guaranteed to be concurrency-safe)
	wmu sync.Mutex // used serialize writes to W (io.Writer is not guaranteed to be concurrency-safe)

	closing   chan struct{} // closing will be closed when the Conn is closed
	closingmu sync.Mutex    // protect closing from concurrent changes (getting set, being closed)

	rdead   time.Time    // read deadline
	rdeadmu sync.RWMutex // lock when being set; rlock when being used

	wdead   time.Time    // write deadline
	wdeadmu sync.RWMutex // lock when being set; rlock when being used
}

// nerr is a Read/Write result
type nerr struct {
	n   int
	err error
}

// Read implements the net.Conn interface.
// Read returns net.OpError errors.
func (c *Conn) Read(b []byte) (int, error) {
	if c.R == nil {
		return 0, c.readErr(false, io.EOF)
	}

	c.initClosing()

	// stop now if we're already closed
	select {
	case <-c.closing:
		return 0, c.readErr(false, connClosed)
	default:
	}

	// stop now if we're already timed out
	c.rdeadmu.RLock()
	defer c.rdeadmu.RUnlock()
	if !c.rdead.IsZero() && c.rdead.Before(time.Now()) {
		return 0, c.readErr(true, timedOut)
	}

	// start read
	readc := make(chan nerr, 1)
	go func() {
		c.rmu.Lock()
		n, err := c.R.Read(b)
		c.rmu.Unlock()
		readc <- nerr{n, err}
	}()

	// set up deadline timeout
	var timeout <-chan time.Time
	timer := deadlineTimer(c.rdead) // c.rdeadmu read lock already held above
	if timer != nil {
		timeout = timer.C
		defer timer.Stop()
	}

	// wait for read success, timeout, or closing
	select {
	case <-c.closing:
		return 0, c.readErr(false, connClosed)
	case <-timeout:
		return 0, c.readErr(true, timedOut)
	case nerr := <-readc:
		if nerr.err != nil {
			// wrap the error
			return nerr.n, c.readErr(false, nerr.err)
		}
		return nerr.n, nil
	}
}

// Write implements the net.Conn interface.
// Write returns net.OpError errors.
func (c *Conn) Write(b []byte) (int, error) {
	if c.W == nil {
		// all writes to Discard succeed, so there's no need to wrap errors
		return ioutil.Discard.Write(b)
	}

	c.initClosing()

	// stop now if we're already closed
	select {
	case <-c.closing:
		return 0, c.writeErr(false, connClosed)
	default:
	}

	// stop now if we're already timed out
	c.wdeadmu.RLock()
	defer c.wdeadmu.RUnlock()
	if !c.wdead.IsZero() && c.wdead.Before(time.Now()) {
		return 0, c.writeErr(true, timedOut)
	}

	// start write
	writec := make(chan nerr, 1)
	go func() {
		c.wmu.Lock()
		n, err := c.W.Write(b)
		c.wmu.Unlock()
		writec <- nerr{n, err}
	}()

	// set up deadline timeout
	var timeout <-chan time.Time
	c.wdeadmu.RLock()
	timer := deadlineTimer(c.wdead) // c.wdeadmu read lock already held above
	if timer != nil {
		timeout = timer.C
		defer timer.Stop()
	}

	// wait for write success, timeout, or closing
	select {
	case <-c.closing:
		return 0, c.writeErr(false, connClosed)
	case <-timeout:
		return 0, c.writeErr(true, timedOut)
	case nerr := <-writec:
		if nerr.err != nil {
			return nerr.n, c.writeErr(false, nerr.err) // wrap the error
		}
		return nerr.n, nil
	}
}

// Close implements the net.Conn interface.
// Closing an already closed Conn will
// return an error (a net.OpError).
func (c *Conn) Close() error {
	c.initClosing()

	c.closingmu.Lock()
	defer c.closingmu.Unlock()

	// short-circuit with an error if we are already closed
	select {
	case <-c.closing:
		return &net.OpError{
			Op:   "close",
			Net:  network,
			Addr: c.LocalAddr(),
			Err:  neterr{timeout: false, err: connClosed},
		}
	default:
		close(c.closing)
	}

	return nil
}

// Wait blocks until the Conn is closed.
func (c *Conn) Wait() {
	c.initClosing()
	<-c.closing
}

// SetDeadline implements the net.Conn interface.
func (c *Conn) SetDeadline(t time.Time) error {
	c.SetReadDeadline(t)
	c.SetWriteDeadline(t)
	return nil
}

// SetReadDeadline implements the net.Conn interface.
func (c *Conn) SetReadDeadline(t time.Time) error {
	c.rdeadmu.Lock()
	c.rdead = t
	c.rdeadmu.Unlock()
	return nil
}

// SetWriteDeadline implements the net.Conn interface.
func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.wdeadmu.Lock()
	c.wdead = t
	c.wdeadmu.Unlock()
	return nil
}

// LocalAddr implements the net.Conn interface.
func (c *Conn) LocalAddr() net.Addr { return addr("local") }

// RemoteAddr implements the net.Conn interface.
func (c *Conn) RemoteAddr() net.Addr { return addr("remote") }

// deadlineTimer returns a time.Timer that fires at the
// provided deadline. If the deadline is 0, it returns nil.
func deadlineTimer(t time.Time) *time.Timer {
	if t.IsZero() {
		return nil
	}
	return time.NewTimer(t.Sub(time.Now()))
}

// initClosing lazily initializes c.closing.
// This helps with the bookkeeping needed
// to make the zero value of Conn usable.
func (c *Conn) initClosing() {
	c.closingmu.Lock()
	if c.closing == nil {
		c.closing = make(chan struct{})
	}
	c.closingmu.Unlock()
}

// readErr wraps a read error in a net.OpError.
func (c *Conn) readErr(timeout bool, e error) *net.OpError {
	return &net.OpError{
		Op:   "read",
		Net:  network,
		Addr: c.RemoteAddr(),
		Err:  neterr{timeout: timeout, err: e},
	}
}

// writeErr wraps a write error in a net.OpError.
func (c *Conn) writeErr(timeout bool, e error) *net.OpError {
	return &net.OpError{
		Op:   "write",
		Net:  network,
		Addr: c.RemoteAddr(),
		Err:  neterr{timeout: timeout, err: e},
	}
}

// Listener is a net.Listener that accepts Conn connections.
type Listener struct {
	sync.Mutex               // hold when mutating any listener state
	connc      chan *Conn    // channel of available connections
	closing    chan struct{} // closing will be closed when the listener is closed
}

// init lazily initializes a Listener.
func (l *Listener) init() {
	l.Lock()
	defer l.Unlock()
	if l.closing == nil {
		l.closing = make(chan struct{})
	}

	// Initialize l.connc only if the listener is not yet closed.
	select {
	case <-l.closing:
	default:
		if l.connc == nil {
			l.connc = make(chan *Conn)
		}
	}
}

// Accept implements the net.Listener interface.
// Accept returns net.OpError errors.
func (l *Listener) Accept() (net.Conn, error) {
	l.init()
	select {
	case conn := <-l.connc:
		return conn, nil
	case <-l.closing:
		operr := &net.OpError{
			Op:   "accept",
			Net:  network,
			Addr: l.Addr(),
			Err:  neterr{timeout: false, err: listenerClosed},
		}
		return nil, operr
	}
}

// Close implements the net.Listener interface.
// Closing an already closed Listener will
// return an error (a net.OpError).
func (l *Listener) Close() error {
	l.init()

	l.Lock()
	defer l.Unlock()

	// short-circuit with an error if we are already closed
	select {
	case <-l.closing:
		return &net.OpError{
			Op:   "close",
			Net:  network,
			Addr: l.Addr(),
			Err:  neterr{timeout: false, err: listenerClosed},
		}
	default:
		l.connc = nil
		close(l.closing)
	}

	return nil
}

// Dial connects to a Listener. r and w may be nil; see Conn.
// Note that r and w here are named from the server's perspective,
// so data that you are sending across the connection will be read
// from r, and responses from the connection will be written to w.
// See the documentation in Conn.
func (l *Listener) Dial(r io.Reader, w io.Writer) (*Conn, error) {
	l.init()
	c := &Conn{R: r, W: w}
	select {
	case <-l.closing:
		operr := &net.OpError{
			Op:   "dial",
			Net:  network,
			Addr: l.Addr(),
			Err:  neterr{timeout: false, err: listenerClosed},
		}
		return nil, operr
	case l.connc <- c:
		return c, nil
	}
}

// Addr implements the net.Listener interface.
func (l *Listener) Addr() net.Addr { return addr("local") }

// addr is a trivial net.Addr
type addr string

func (a addr) Network() string { return network }
func (a addr) String() string  { return string(a) }

const network = "ionet"

// neterr is a simple net.Error
type neterr struct {
	temporary bool
	timeout   bool
	err       error
}

func (e neterr) Temporary() bool { return e.temporary }
func (e neterr) Timeout() bool   { return e.timeout }
func (e neterr) Error() string   { return e.err.Error() }

var (
	connClosed     = fmt.Errorf("conn closed")
	timedOut       = fmt.Errorf("timed out")
	listenerClosed = fmt.Errorf("listener closed")
)
