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

package processors

import (
	"fmt"
	"sync"

	"github.com/apache/thrift/lib/go/thrift"
	customtransport "github.com/jaegertracing/jaeger/cmd/agent/app/customtransports"
	log "github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/agent/jaegeragent/thrift-udp/servers"
)

// ThriftProcessor is a server that processes spans using a TBuffered Server
type ThriftProcessor struct {
	server        servers.Server
	handler       AgentProcessor
	protocolPool  *sync.Pool
	numProcessors int
	processing    sync.WaitGroup
}

// AgentProcessor handler used by the processor to process thrift and call the reporter with the deserialized struct
type AgentProcessor interface {
	Process(iprot, oprot thrift.TProtocol) (success bool, err thrift.TException)
}

// NewThriftProcessor creates a TBufferedServer backed ThriftProcessor
func NewThriftProcessor(
	server servers.Server,
	numProcessors int,
	factory thrift.TProtocolFactory,
	handler AgentProcessor,
) (*ThriftProcessor, error) {
	if numProcessors <= 0 {
		return nil, fmt.Errorf(
			"Number of processors must be greater than 0, called with %d", numProcessors)
	}
	var protocolPool = &sync.Pool{
		New: func() interface{} {
			trans := &customtransport.TBufferedReadTransport{}
			return factory.GetProtocol(trans)
		},
	}

	res := &ThriftProcessor{
		server:        server,
		handler:       handler,
		protocolPool:  protocolPool,
		numProcessors: numProcessors,
	}
	return res, nil
}

// Serve initiates the readers and starts serving traffic
func (s *ThriftProcessor) Serve() {
	s.processing.Add(s.numProcessors)
	for i := 0; i < s.numProcessors; i++ {
		go s.processBuffer()
	}

	s.server.Serve()
}

// IsServing indicates whether the server is currently serving traffic
func (s *ThriftProcessor) IsServing() bool {
	return s.server.IsServing()
}

// Stop stops the serving of traffic and waits until the queue is
// emptied by the readers
func (s *ThriftProcessor) Stop() {
	s.server.Stop()
	s.processing.Wait()
}

// processBuffer reads data off the channel and puts it into a custom transport for
// the processor to process
func (s *ThriftProcessor) processBuffer() {
	for readBuf := range s.server.DataChan() {
		protocol := s.protocolPool.Get().(thrift.TProtocol)
		payload := readBuf.GetBytes()
		protocol.Transport().Write(payload)
		s.server.DataRecd(readBuf) // acknowledge receipt and release the buffer

		if ok, _ := s.handler.Process(protocol, protocol); !ok {
			log.Print()
		}
		s.protocolPool.Put(protocol)
	}
	s.processing.Done()
}
