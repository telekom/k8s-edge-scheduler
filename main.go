// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package main

import (
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/kubeclient"
	"github.com/telekom/k8s-edge-scheduler/scheduler"
)

func main() {
	debug := flag.Bool("debug", false, "enable debugging")
	flag.Parse()

	log := logrus.New()
	if *debug {
		log.SetLevel(logrus.DebugLevel)
	}

	kube := kubeclient.NewKubeClient(log, kubeclient.NewClientset())

	s := scheduler.NewScheduler(kube, log)
	s.Start()
}
