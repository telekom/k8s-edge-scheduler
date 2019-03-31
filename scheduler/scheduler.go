// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package scheduler

import (
	"time"

	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/cache"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware/deploymentstatus"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware/location"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware/nodeselector"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	watch "k8s.io/client-go/tools/cache"
)

var (
	log                *logrus.Entry
	name               string
	namespace          string
	descheduleInterval time.Duration
)

func init() {
	flag.StringVar(&name, "name", "edge-scheduler", "scheduler name")
	flag.StringVar(&namespace, "namespace", "default", "kubernetes namespace")
	flag.DurationVar(&descheduleInterval, "descheduleInterval", time.Minute, "interval to check pods for descheduling")
}

type Scheduler struct {
	name               string
	namespace          string
	kube               KubernetesClient
	nodes              *cache.Cache
	scheduleM          middleware.Middleware
	descheduleM        middleware.Middleware
	descheduleInterval time.Duration
	decisions          *cache.Cache
}

type KubernetesClient interface {
	GetClientset() *kubernetes.Clientset
	GetDeploymentFromPod(p *v1.Pod) (*appsv1.Deployment, error)
}

func NewScheduler(k KubernetesClient, l *logrus.Logger) *Scheduler {
	log = l.WithFields(logrus.Fields{
		"component": "scheduler",
	})

	s := &Scheduler{
		name:               name,
		namespace:          namespace,
		kube:               k,
		nodes:              cache.NewCache(),
		descheduleInterval: descheduleInterval,
		decisions:          cache.NewCache(),
	}
	s.nodes.Timeout = 0 * time.Second
	s.decisions.Timeout = 0 * time.Second

	return s
}

func (s *Scheduler) Start() {
	s.scheduleM = middleware.Adapt(
		s.bindPod,
		nodeselector.NodeSelector,
		location.Location,
		deploymentstatus.DeploymentStatus,
	)

	s.descheduleM = middleware.Adapt(
		s.evictPod,
		nodeselector.NodeSelector,
		location.Location,
		deploymentstatus.DeploymentStatus,
	)

	s.watchNodes()

	podList := watch.NewListWatchFromClient(s.kube.GetClientset().CoreV1().RESTClient(), "pods", s.namespace, fields.Everything())
	_, controller := watch.NewInformer(podList, &v1.Pod{}, time.Second*0, watch.ResourceEventHandlerFuncs{
		AddFunc: s.schedule,
	})
	log.Infof("watch as %s for new pods in namespace %s", s.name, s.namespace)
	go controller.Run(make(chan struct{}))

	for {
		s.deschedule()
		<-time.NewTimer(s.descheduleInterval).C
	}
}

func (s *Scheduler) GetKube() middleware.KubernetesClient {
	return s.kube.(middleware.KubernetesClient)
}

func (s *Scheduler) GetNamespace() string {
	return s.namespace
}

func (s *Scheduler) GetNodes() *cache.Cache {
	return s.nodes
}

func (s *Scheduler) Log(component string) *logrus.Entry {
	return log.WithFields(logrus.Fields{
		"component": component + "-middleware",
	})
}
