package ionet

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"
)

// timedWaitGroup extends sync.WaitGroup with an option
// to wait for a limited time.
type timedWaitGroup struct{ sync.WaitGroup }

// WaitFor blocks until the WaitGroup counter is zero or duration d
// has elapsed. It returns true iff it was unblocked due to the counter
// dropping to zero.
func (t *timedWaitGroup) WaitFor(d time.Duration) (completed bool) {
	timeout := time.After(d)
	success := make(chan struct{})
	go func() {
		t.WaitGroup.Wait()
		close(success)
	}()
	select {
	case <-success:
		return true
	case <-timeout:
		return false
	}
}

func TestConnBasics(t *testing.T) {
	in := []byte("you can't read from a closed connection (but you can be happy, if you've a mind to)")
	r := bytes.NewBuffer(in)
	w := new(bytes.Buffer)
	c := &Conn{R: r, W: w}
	b := make([]byte, len(in))

	if c.LocalAddr() == nil || c.RemoteAddr() == nil {
		t.Errorf("conn should have a local and remote addr")
	}

	n, err := c.Read(b)
	if n == 0 || err != nil {
		t.Errorf("conn read failed with err %v", err)
	}
	if !bytes.Equal(b, in) {
		t.Errorf("read incorrect bytes, want %q got %q", in, b)
	}

	n, err = c.Write(b)
	if n == 0 || err != nil {
		t.Errorf("conn write failed with err %v", err)
	}
	if !bytes.Equal(w.Bytes(), in) {
		t.Errorf("wrote incorrect bytes, want %q got %q", w.Bytes(), in)
	}

	err = c.Close()
	if n == 0 || err != nil {
		t.Errorf("close err %v", err)
	}
	n, err = c.Read(b)
	if n != 0 || err == nil {
		t.Errorf("conn read should fail after close, instead read %d bytes", n)
	}
	n, err = c.Write(b)
	if n != 0 || err == nil {
		t.Errorf("conn write should fail after close, instead read %d bytes", n)
	}
}

func TestConnDoubleClose(t *testing.T) {
	c := &Conn{}
	err := c.Close()
	if err != nil {
		t.Errorf("conn close should succeed the first time, got error: %s", err)
	}
	err = c.Close()
	if err == nil {
		t.Errorf("conn close should fail with an error the second time")
	} else if err := err.(*net.OpError); err.Temporary() {
		t.Errorf("conn close should fail with a permanent error the second time")
	}
}

func TestConnZero(t *testing.T) {
	c := new(Conn)
	b := []byte{0x00}
	n, err := c.Read(b)
	if n != 0 || err == nil {
		t.Errorf("zero conn reads should return 0, wrapped EOF, got %d, %v", n, err)
	}
	n, err = c.Write(b)
	if n != len(b) || err != nil {
		t.Errorf("zero conn writes should succeed after writing %d bytes, wrote %d with err %v", len(b), n, err)
	}
	err = c.Close()
	if err != nil {
		t.Errorf("zero conn should close without error, got %v", err)
	}
}

// blockrw is an io.ReadWriter that blocks forever on reads and writes.
type blockrw struct{}

func (*blockrw) Read(b []byte) (int, error)  { select {} }
func (*blockrw) Write(b []byte) (int, error) { select {} }

func TestConnCloseUnblocksReadWrites(t *testing.T) {
	c := &Conn{R: &blockrw{}, W: &blockrw{}}
	b := []byte{}
	var w timedWaitGroup
	w.Add(2)
	go func() {
		c.Read(b)
		w.Done()
	}()
	go func() {
		c.Write(b)
		w.Done()
	}()
	time.Sleep(time.Millisecond * 5) // ensure that the read/write has started
	c.Close()
	c.Wait()
	if !w.WaitFor(time.Millisecond * 10) {
		t.Errorf("conn close did not unblock read/write")
	}
}

func TestConnTimeouts(t *testing.T) {
	c := &Conn{R: &blockrw{}, W: &blockrw{}}
	c.SetDeadline(time.Now().Add(time.Millisecond * 10))
	b := []byte{}
	var w timedWaitGroup
	w.Add(2)
	go func() {
		c.Read(b)
		w.Done()
	}()
	go func() {
		c.Write(b)
		w.Done()
	}()
	if !w.WaitFor(time.Millisecond * 50) {
		t.Errorf("conn timeout did not unblock read/write (deadline after 10ms, waited 50ms)")
	}
}

// errorrw is an io.ReadWriter that fails tests when read from or written to.
type errorrw struct {
	t   *testing.T
	msg string
}

func (e *errorrw) Read(b []byte) (int, error) {
	e.t.Errorf("unexpected read: %v", e.msg)
	return 0, fmt.Errorf("errorrw does not read")
}

func (e *errorrw) Write(b []byte) (int, error) {
	e.t.Errorf("unexpected write: %v", e.msg)
	return 0, fmt.Errorf("errorrw does not write")
}

func TestConnNoReadWriteAfterClose(t *testing.T) {
	erw := &errorrw{t: t, msg: "no reads or writes after connection close"}
	b := []byte{}
	c := &Conn{R: erw, W: erw}
	c.Close()
	c.Read(b)
	c.Write(b)
}

func TestConnNoReadWriteAfterTimeout(t *testing.T) {
	erw := &errorrw{t: t, msg: "no reads or writes after timeout"}
	b := []byte{}
	c := &Conn{R: erw, W: erw}
	c.SetDeadline(time.Now().Add(-time.Second))
	_, err := c.Read(b)
	if err := err.(*net.OpError); !err.Timeout() {
		t.Errorf("read timeout errors should be marked as timeouts")
	}
	_, err = c.Write(b)
	if err := err.(*net.OpError); !err.Timeout() {
		t.Errorf("write timeout errors should be marked as timeouts")
	}
}

// seqrw is an io.ReadWriter that does not handle concurrency.
// It panics when multiple concurrent reads/writes occur.
// And it reads and writes slowly, to provide ample opportunity
// for concurrent access.
type seqrw struct {
	reading bool
	writing bool
	t       *testing.T
	d       time.Duration
}

func (s *seqrw) Read(b []byte) (int, error) {
	if s.reading {
		s.t.Errorf("non-sequential read")
	}
	s.reading = true
	time.Sleep(s.d)
	s.reading = false
	return 0, fmt.Errorf("seqrw does not read")
}

func (s *seqrw) Write(b []byte) (int, error) {
	if s.writing {
		s.t.Errorf("non-sequential write")
	}
	s.writing = true
	time.Sleep(s.d)
	s.writing = false
	return 0, fmt.Errorf("seqrw does not write")
}

func TestConnEnforcesSequentialReadWrites(t *testing.T) {
	s := &seqrw{t: t, d: time.Millisecond * 10}
	b := []byte{0x00}
	c := &Conn{R: s, W: s}
	var w sync.WaitGroup
	for i := 0; i < 2; i++ {
		w.Add(1)
		go func() {
			c.Read(b)
			w.Done()
		}()
		w.Add(1)
		go func() {
			c.Write(b)
			w.Done()
		}()
	}
	w.Wait()
}

func TestListenerAcceptsWhenDialed(t *testing.T) {
	l := Listener{}
	timeout := time.After(time.Millisecond * 10)
	successc := make(chan struct{}, 1)
	go func() {
		conn, err := l.Accept()
		if conn == nil || err != nil {
			t.Errorf("accept error: %v", err)
		}
		successc <- struct{}{}
	}()
	go func() {
		l.Dial(nil, nil)
	}()
	select {
	case <-successc:
	case <-timeout:
		t.Errorf("listener was dialed but failed to accept")
	}
}

func TestListenerAcceptManyConcurrently(t *testing.T) {
	l := Listener{}
	n := 10
	wg := sync.WaitGroup{}
	timeout := time.After(time.Millisecond * 25) // enough time for 10 concurrent 10ms reads/writes, but not enough to 10 serial ones
	successc := make(chan struct{}, 1)
	// Fire off a bunch of accepts
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			conn, err := l.Accept()
			if conn == nil || err != nil {
				t.Errorf("accept error: %v", err)
			}
			b := []byte{0x00}
			conn.Read(b)
			conn.Write(b)
			wg.Done()
		}()
	}
	// Fire off a bunch of dials with slow read/writers
	for i := 0; i < n; i++ {
		s := &seqrw{t: t, d: time.Millisecond * 5}
		go func() {
			l.Dial(s, s)
		}()
	}
	// Wait for all reads and writes to complete;
	// this will happen in time iff they are concurrent.
	go func() {
		wg.Wait()
		successc <- struct{}{}
	}()
	select {
	case <-successc:
	case <-timeout:
		t.Errorf("listener did not accept concurrently (too slow)")
	}
}

func TestListenerDialFailsWhenClosed(t *testing.T) {
	l := Listener{}
	l.Close()
	conn, err := l.Dial(nil, nil)
	if conn != nil || err == nil {
		t.Errorf("dialing a listener should fail with an error when the listener is closed")
	}
}

func TestListenerCloseUnblocksAccept(t *testing.T) {
	l := Listener{}
	var w timedWaitGroup
	w.Add(1)
	go func() {
		l.Accept()
		w.Done()
	}()
	l.Close()
	if !w.WaitFor(time.Millisecond * 10) {
		t.Errorf("listener close did not unblock accept")
	}
}

func TestListenerDoubleClose(t *testing.T) {
	l := &Listener{}
	err := l.Close()
	if err != nil {
		t.Errorf("listener close should succeed the first time, got error: %s", err)
	}
	err = l.Close()
	if err == nil {
		t.Errorf("listener close should fail with an error the second time")
	} else if err := err.(*net.OpError); err.Temporary() {
		t.Errorf("listener close should fail with a permanent error the second time")
	}
}

// When you're at 9x% test coverage, and the remaining lines are trivial, it's so
// very, very hard to resist the temptation to push to 100%. :)
func TestNothingAtAllImportant(t *testing.T) {
	a := addr("addr")
	if a.Network() == "" || a.String() == "" {
		t.Errorf("addrs should have a non-empty network and string description")
	}

	s := "dummy err"
	e := neterr{err: fmt.Errorf(s)}
	if e.Error() != s {
		t.Errorf("err did not return the wrapped error")
	}
}
