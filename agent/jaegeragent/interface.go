// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package jaegeragent

import (
	"fmt"
	"net/url"

	"github.com/jaegertracing/jaeger/thrift-gen/jaeger"
	"github.com/jaegertracing/jaeger/thrift-gen/zipkincore"
)

func (j *JaegerAgent) EmitBatch(batch *jaeger.Batch) error {
	log.Debugf("received batch from %s", batch.Process.GetServiceName())

	var hostname string
	for _, tag := range batch.Process.GetTags() {
		if tag.GetKey() == "hostname" {
			hostname = tag.GetVStr()
			break
		}
	}

	for _, span := range j.filterSpansByType(batch.GetSpans(), "client") {
		for _, tag := range span.GetTags() {
			if tag.GetKey() == "http.url" {
				url, err := url.Parse(tag.GetVStr())
				if err != nil {
					log.Warnf("destination is not an url: %s", tag.GetVStr())
				} else {
					j.collector.AddRequest(url.Hostname(), batch.Process.GetServiceName(), hostname, span.GetStartTime(), span.GetDuration())
				}
			}
		}
	}
	return nil
}

func (j *JaegerAgent) EmitZipkinBatch([]*zipkincore.Span) error {
	log.Println("zipkin batch not implemented")
	return fmt.Errorf("not implemented")
}

func (j *JaegerAgent) filterSpansByType(spans []*jaeger.Span, spanType string) []*jaeger.Span {
	var res []*jaeger.Span
	for _, span := range spans {
		for _, tag := range span.GetTags() {
			if tag.GetKey() == "span.kind" && tag.IsSetVStr() && tag.GetVStr() == spanType {
				res = append(res, span)
			}
		}
	}
	return res
}
