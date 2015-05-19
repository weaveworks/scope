package xfer

import (
	"encoding/gob"
	"io"
	"log"
	"net"

	"github.com/weaveworks/scope/report"
)

// Publisher provides a way to send reports upstream.
type Publisher interface {
	Publish(report.Report)
	Close()
}

// TCPPublisher is a Publisher implementation which uses TCP and gob encoding.
type TCPPublisher struct {
	msg    chan report.Report
	closer io.Closer
}

// NewTCPPublisher listens for connections on listenAddress. Only one client
// is accepted at a time; other clients are accepted, but disconnected right
// away. Reports published via publish() will be written to the connected
// client, if any. Gentle shutdown of the returned publisher via close().
func NewTCPPublisher(listenAddress string) (*TCPPublisher, error) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, err
	}

	p := &TCPPublisher{
		msg:    make(chan report.Report),
		closer: listener,
	}

	go p.loop(fwd(listener))

	return p, nil
}

// Close stops a TCPPublisher and closes the socket.
func (p *TCPPublisher) Close() {
	close(p.msg)
	p.closer.Close()
}

// Publish sens a Report to the client, if any.
func (p *TCPPublisher) Publish(msg report.Report) {
	p.msg <- msg
}

func (p *TCPPublisher) loop(incoming <-chan net.Conn) {
	activeConns := map[string]connEncoder{} // host: connEncoder

	for {
		select {
		case conn, ok := <-incoming:
			if !ok {
				return // someone closed our connection chan -- weird?
			}

			// Don't allow multiple connections from the same remote host.
			host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
			if err != nil {
				log.Printf("incoming connection: %s: %v (dropped)", conn.RemoteAddr(), err)
				conn.Close()
				continue
			}
			if _, ok := activeConns[host]; ok {
				log.Printf("duplicate connection: %s (dropped)", conn.RemoteAddr())
				conn.Close()
				continue
			}

			log.Printf("connection initiated: %s", conn.RemoteAddr())
			activeConns[host] = connEncoder{conn, gob.NewEncoder(conn)}

		case msg, ok := <-p.msg:
			if !ok {
				return // someone closed our msg chan, so we're done
			}

			for host, connEncoder := range activeConns {
				if err := connEncoder.Encoder.Encode(msg); err != nil {
					log.Printf("connection terminated: %s: %v", connEncoder.Conn.RemoteAddr(), err)
					connEncoder.Conn.Close()
					delete(activeConns, host)
				}
			}
		}
	}
}

type connEncoder struct {
	net.Conn
	*gob.Encoder
}

func fwd(ln net.Listener) chan net.Conn {
	c := make(chan net.Conn)

	go func() {
		defer close(c)
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("%s: %s", ln.Addr(), err)
				return
			}
			c <- conn
		}
	}()

	return c
}
