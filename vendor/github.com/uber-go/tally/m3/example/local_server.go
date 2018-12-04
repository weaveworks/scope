// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package main

import (
	"bytes"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/uber-go/tally/m3"
	customtransport "github.com/uber-go/tally/m3/customtransports"
	m3thrift "github.com/uber-go/tally/m3/thrift"

	"github.com/apache/thrift/lib/go/thrift"
)

type batchCallback func(batch *m3thrift.MetricBatch)

type localM3Server struct {
	Service   *localM3Service
	Addr      string
	protocol  m3.Protocol
	processor thrift.TProcessor
	conn      *net.UDPConn
	closed    int32
}

func newLocalM3Server(
	listenAddr string,
	protocol m3.Protocol,
	fn batchCallback,
) (*localM3Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	service := newLocalM3Service(fn)
	processor := m3thrift.NewM3Processor(service)
	conn, err := net.ListenUDP(udpAddr.Network(), udpAddr)
	if err != nil {
		return nil, err
	}

	return &localM3Server{
		Service:   service,
		Addr:      conn.LocalAddr().String(),
		conn:      conn,
		protocol:  protocol,
		processor: processor,
	}, nil
}

func (f *localM3Server) Serve() error {
	readBuf := make([]byte, 65536)
	for {
		n, err := f.conn.Read(readBuf)
		if err != nil {
			if atomic.LoadInt32(&f.closed) == 0 {
				return fmt.Errorf("failed to read: %v", err)
			}
			return nil
		}
		trans, _ := customtransport.NewTBufferedReadTransport(bytes.NewBuffer(readBuf[0:n]))
		var proto thrift.TProtocol
		if f.protocol == m3.Compact {
			proto = thrift.NewTCompactProtocol(trans)
		} else {
			proto = thrift.NewTBinaryProtocolTransport(trans)
		}
		f.processor.Process(proto, proto)
	}
}

func (f *localM3Server) Close() error {
	atomic.AddInt32(&f.closed, 1)
	return f.conn.Close()
}

type localM3Service struct {
	fn batchCallback
}

func newLocalM3Service(fn batchCallback) *localM3Service {
	return &localM3Service{fn: fn}
}

func (m *localM3Service) EmitMetricBatch(batch *m3thrift.MetricBatch) (err error) {
	m.fn(batch)
	return thrift.NewTTransportException(thrift.END_OF_FILE, "complete")
}
