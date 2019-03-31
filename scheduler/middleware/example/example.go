// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package example

import (
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
)

const name = "example"

var (
	log *logrus.Entry
)

func Example(m middleware.Middleware) middleware.Middleware {
	return func(s middleware.Scheduler, d *middleware.Data) {
		log = s.Log(name)
		defer m(s, d)
		// log.Debug("hello from example")
	}
}
