// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package scheduler

import (
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
	"github.com/telekom/k8s-edge-scheduler/scheduler/priomap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scheduler) schedule(obj interface{}) {
	pod := obj.(*v1.Pod)
	if pod.Status.Phase == "Pending" && pod.Spec.SchedulerName == s.name {
		d, err := s.kube.GetDeploymentFromPod(pod)
		if err != nil {
			log.Warnf("pod %s belongs to no deployment", pod.Name)
			return
		}

		data := &middleware.Data{
			Pod:        pod,
			Deployment: d,
		}

		if m, ok := s.decisions.Get(d.Name); ok {
			s.decisions.Delete(d.Name)
			log.Debugf("found decision for deployment %s in cache", d.Name)
			data.Prio = m.(*middleware.Data).Prio
			s.bindPod(nil, data)
			return
		}

		data.Prio = priomap.NewNodePrioMap(s.nodes.Keys())
		s.scheduleM(s, data)
	}
}

func (s *Scheduler) bindPod(_ middleware.Scheduler, d *middleware.Data) {
	node := d.Prio.(*priomap.NodePrioMap).Max()
	if node == "" {
		log.Warnf("no node found for pod %s", d.Pod.Name)
		return
	}

	// bind pod
	err := s.kube.GetClientset().CoreV1().Pods(s.namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Pod.Name,
		},
		Target: v1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Node",
			Name:       node,
		},
	})
	if err != nil {
		log.Warnf("could not bind pod %s to node %s: %s", d.Pod.Name, node, err.Error())
	} else {
		log.Infof("bind pod %s to node %s", d.Pod.Name, node)
	}
}
