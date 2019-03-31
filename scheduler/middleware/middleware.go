// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package middleware

import (
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/cache"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Middleware func(scheduler Scheduler, data *Data)

type Adapter func(Middleware) Middleware

type Scheduler interface {
	GetKube() KubernetesClient
	GetNodes() *cache.Cache
	GetNamespace() string
	Log(component string) *logrus.Entry
}

type KubernetesClient interface {
	GetClientset() *kubernetes.Clientset
	GetPodsFromDeployment(d *appsv1.Deployment) ([]v1.Pod, error)
	GetLocationFromNode(n *v1.Node) (string, error)
}

type PrioMap interface {
	Get(k string) (int, error)
	Add(k string, add int) error
	Set(k string, v int) error
	Disable(k string) error
	Keys() []string
	ListInc() []PrioMapPair
	ListDec() []PrioMapPair
}

type Data struct {
	Pod        *v1.Pod
	Deployment *appsv1.Deployment
	Prio       PrioMap
}

type PrioMapPair struct {
	Key   string
	Value int
}

func Adapt(m Middleware, adapters ...Adapter) Middleware {
	for i := range adapters {
		m = adapters[len(adapters)-1-i](m)
	}
	return m
}
