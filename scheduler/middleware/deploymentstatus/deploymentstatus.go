// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package deploymentstatus

import (
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
	v1 "k8s.io/api/core/v1"
)

const name = "deploymentstatus"

var (
	log     *logrus.Entry
	maxPods int
)

func init() {
	flag.IntVar(&maxPods, "maxPods", 2, "max pods per node")
}

func DeploymentStatus(m middleware.Middleware) middleware.Middleware {
	return func(s middleware.Scheduler, d *middleware.Data) {
		log = s.Log(name)
		defer m(s, d)

		podCount, err := getPodCountByNode(s, d)
		if err != nil {
			log.Warn(err.Error())
			return
		}

		for _, n := range d.Prio.Keys() {
			if c, ok := podCount[n]; ok && c > 0 {
				log.Debugf("node %s runs %d other pods of deployment %s", n, c, d.Deployment.Name)
				if c < maxPods {
					o, _ := d.Prio.Get(n)
					d.Prio.Set(n, o/(c+1))
				} else {
					d.Prio.Disable(n)
				}
			}
		}
	}
}

func getPodCountByNode(s middleware.Scheduler, d *middleware.Data) (map[string]int, error) {
	other, err := s.GetKube().GetPodsFromDeployment(d.Deployment)
	if err != nil {
		return nil, err
	}

	r := make(map[string]int)
	for _, n := range d.Prio.Keys() {
		r[n] = 0
	}
	for _, p := range other {
		if d.Pod.Name != p.Name && isReady(&p) {
			if _, ok := r[p.Spec.NodeName]; ok {
				r[p.Spec.NodeName]++
			}
		}
	}

	return r, nil
}

func isReady(p *v1.Pod) bool {
	if p.Status.Phase == v1.PodRunning {
		return true
	}
	for _, c := range p.Status.Conditions {
		if c.Type == v1.PodReady || c.Type == v1.PodInitialized || c.Type == v1.PodScheduled {
			return true
		}
	}
	return false
}
