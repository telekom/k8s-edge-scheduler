// Copyright (c) 2017 Uber Technologies, Inc.
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

// modified by Lukas Steiner for logrus logging

package thriftudp

import (
	"bytes"
	"errors"
	"net"
	"sync/atomic"

	log "github.com/sirupsen/logrus"

	"github.com/apache/thrift/lib/go/thrift"
)

//MaxLength of UDP packet
const MaxLength = 65000

var errConnAlreadyClosed = errors.New("connection already closed")

// TUDPTransport does UDP as a thrift.TTransport
type TUDPTransport struct {
	conn     *net.UDPConn
	addr     *net.UDPAddr
	writeBuf bytes.Buffer
	closed   uint32 // atomic flag
}

// NewTUDPServerTransport creates a net.UDPConn-backed TTransport for Thrift servers
// It will listen for incoming udp packets on the specified host/port
// Example:
// 	trans, err := thriftudp.NewTUDPClientTransport("localhost:9001")
func NewTUDPServerTransport(hostPort string) (*TUDPTransport, error) {
	addr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return nil, thrift.NewTTransportException(thrift.NOT_OPEN, err.Error())
	}
	conn, err := net.ListenUDP(addr.Network(), addr)
	if err != nil {
		return nil, thrift.NewTTransportException(thrift.NOT_OPEN, err.Error())
	}

	return &TUDPTransport{
		addr: addr,
		conn: conn,
	}, nil
}

func (p *TUDPTransport) Listen() error {
	return nil
}

func (p *TUDPTransport) Accept() (thrift.TTransport, error) {
	log.Print("accept called")
	return p, nil
}

func (p *TUDPTransport) Interrupt() error {
	log.Print("interrupt called")
	return nil
}

// Open does nothing as connection is opened on creation
// Required to maintain thrift.TTransport interface
func (p *TUDPTransport) Open() error {
	return nil
}

// Conn retrieves the underlying net.UDPConn
func (p *TUDPTransport) Conn() *net.UDPConn {
	return p.conn
}

// IsOpen returns true if the connection is open
func (p *TUDPTransport) IsOpen() bool {
	return atomic.LoadUint32(&p.closed) == 0
}

// Close closes the connection
func (p *TUDPTransport) Close() error {
	if atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		return p.conn.Close()
	}
	return errConnAlreadyClosed
}

// Addr returns the address that the transport is listening on or writing to
func (p *TUDPTransport) Addr() net.Addr {
	return p.addr
}

// Read reads one UDP packet and puts it in the specified buf
func (p *TUDPTransport) Read(buf []byte) (int, error) {
	if !p.IsOpen() {
		return 0, thrift.NewTTransportException(thrift.NOT_OPEN, "Connection not open")
	}
	n, err := p.conn.Read(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

// RemainingBytes returns the max number of bytes (same as Thrift's StreamTransport) as we
// do not know how many bytes we have left.
func (p *TUDPTransport) RemainingBytes() uint64 {
	const maxSize = ^uint64(0)
	return maxSize
}

// Write writes specified buf to the write buffer
func (p *TUDPTransport) Write(buf []byte) (int, error) {
	if !p.IsOpen() {
		return 0, thrift.NewTTransportException(thrift.NOT_OPEN, "Connection not open")
	}
	if len(p.writeBuf.Bytes())+len(buf) > MaxLength {
		return 0, thrift.NewTTransportException(thrift.INVALID_DATA, "Data does not fit within one UDP packet")
	}
	n, err := p.writeBuf.Write(buf)
	return n, thrift.NewTTransportExceptionFromError(err)
}

// Flush flushes the write buffer as one udp packet
func (p *TUDPTransport) Flush() error {
	if !p.IsOpen() {
		return thrift.NewTTransportException(thrift.NOT_OPEN, "Connection not open")
	}

	_, err := p.conn.Write(p.writeBuf.Bytes())
	p.writeBuf.Reset() // always reset the buffer, even in case of an error
	return err
}