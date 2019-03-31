// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package nodeselector

import (
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
	v1 "k8s.io/api/core/v1"
)

func NodeSelector(m middleware.Middleware) middleware.Middleware {
	return func(s middleware.Scheduler, d *middleware.Data) {

		s.GetNodes().Mutex.Lock()
		for _, k := range d.Prio.Keys() {
			o, _ := s.GetNodes().Get(k)
			n := o.(*v1.Node)

			// node selector
			for l, v := range d.Pod.Spec.NodeSelector {
				if a, ok := n.Labels[l]; !ok || a != v {
					d.Prio.Disable(k)
				}
			}

			// tolerations
			for _, t := range n.Spec.Taints {
				for _, l := range d.Pod.Spec.Tolerations {
					if t.Effect == "NoSchedule" &&
						(t.Key != l.Key || t.Value != t.Value) {
						d.Prio.Disable(k)
					}
				}
			}
		}
		s.GetNodes().Mutex.Unlock()
		m(s, d)
	}
}
