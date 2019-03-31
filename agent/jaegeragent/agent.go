// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package jaegeragent

import (
	"github.com/apache/thrift/lib/go/thrift"
	"github.com/jaegertracing/jaeger/thrift-gen/agent"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/agent/jaegeragent/thrift-udp/processors"
	"github.com/telekom/k8s-edge-scheduler/agent/jaegeragent/thrift-udp/servers"
	"github.com/telekom/k8s-edge-scheduler/agent/jaegeragent/thrift-udp/servers/thriftudp"
)

var (
	log  *logrus.Entry
	addr string
)

func init() {
	flag.StringVar(&addr, "jaegerAddr", ":6831", "jaeger agent address")
}

type JaegerAgent struct {
	addr      string
	processor *processors.ThriftProcessor
	collector Collector
}

type Collector interface {
	AddRequest(ip string, proxy string, hostname string, timestamp int64, duration int64)
}

func NewJaegerAgent(c Collector, l *logrus.Logger) *JaegerAgent {
	log = l.WithFields(logrus.Fields{
		"component": "jaegeragent",
	})

	a := &JaegerAgent{
		addr:      addr,
		collector: c,
	}
	protocolFactory := thrift.NewTCompactProtocolFactory()
	transport, err := thriftudp.NewTUDPServerTransport(addr)
	if err != nil {
		log.Fatalf("error while creating jaeger agent transport: %s", err.Error())
	}
	agent := agent.NewAgentProcessor(a)

	server, err := servers.NewTBufferedServer(transport, 1000, 65000)
	if err != nil {
		log.Fatalf("error while creating jaeger agent server: %s", err.Error())
	}
	a.processor, err = processors.NewThriftProcessor(server, 1, protocolFactory, agent)
	if err != nil {
		log.Fatalf("error while creating jaeger agent processor: %s", err.Error())
	}
	return a
}

func (a *JaegerAgent) Serve() {
	log.Infof("listen on %s as jaeger agent", a.addr)
	a.processor.Serve()
}
