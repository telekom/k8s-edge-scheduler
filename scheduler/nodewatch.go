// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package scheduler

import (
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	watch "k8s.io/client-go/tools/cache"
)

func (s *Scheduler) watchNodes() {
	nodeList := watch.NewListWatchFromClient(s.kube.GetClientset().CoreV1().RESTClient(), "nodes", "", fields.Everything())
	_, controller := watch.NewInformer(nodeList, &v1.Node{}, time.Second*0, watch.ResourceEventHandlerFuncs{
		AddFunc:    s.addNode,
		DeleteFunc: s.deleteNode,
	})
	go controller.Run(make(chan struct{}))
}

func (s *Scheduler) addNode(o interface{}) {
	n := o.(*v1.Node)
	log.Debugf("add node %s", n.Name)
	s.nodes.Set(n.Name, n)
}

func (s *Scheduler) deleteNode(o interface{}) {
	n := o.(*v1.Node)
	s.nodes.Delete(n.Name)
	log.Debugf("delete node %s", n.Name)
}
