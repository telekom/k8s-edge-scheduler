// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package scheduler

import (
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
	"github.com/telekom/k8s-edge-scheduler/scheduler/priomap"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (s *Scheduler) deschedule() {
	pods, err := s.kube.GetClientset().CoreV1().Pods(s.namespace).List(metav1.ListOptions{})
	if err != nil {
		log.Warn(err.Error())
	}
	for _, p := range pods.Items {
		if p.Status.Phase == "Running" && p.Spec.SchedulerName == s.name {
			d, err := s.kube.GetDeploymentFromPod(&p)
			if err != nil {
				log.Warnf("pod %s belongs to no deployment", p.Name)
			}
			s.descheduleM(s, &middleware.Data{
				Pod:        &p,
				Deployment: d,
				Prio:       priomap.NewNodePrioMap(s.nodes.Keys()),
			})
		}
	}
}

func (s *Scheduler) evictPod(_ middleware.Scheduler, d *middleware.Data) {
	node := d.Prio.(*priomap.NodePrioMap).Max()
	p, _ := d.Prio.Get(node)

	if node == "" || d.Pod.Spec.NodeName == node || p == 0 {
		log.Debugf("pod %s is placed at the right node", d.Pod.Name)
		return
	}

	e := &policy.Eviction{
		TypeMeta: metav1.TypeMeta{
			APIVersion: d.Pod.APIVersion,
			Kind:       d.Pod.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      d.Pod.Name,
			Namespace: d.Pod.Namespace,
		},
		DeleteOptions: &metav1.DeleteOptions{},
	}

	err := s.kube.GetClientset().Policy().Evictions(d.Pod.Namespace).Evict(e)
	if err != nil {
		log.Warnf("cloud no evict pod %s: %s", d.Pod.Name, err.Error())
		return
	}
	log.Infof("evict pod %s from node %s", d.Pod.Name, d.Pod.Spec.NodeName)

	s.decisions.Set(d.Deployment.Name, d)
}
