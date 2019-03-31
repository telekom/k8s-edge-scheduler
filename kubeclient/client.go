// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package kubeclient

import (
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	log        *logrus.Entry
	kubeConfig string
)

func init() {
	flag.StringVar(&kubeConfig, "kubeConfig", "incluster", "kubernetes config")
}

type KubeClient struct {
	clientset *kubernetes.Clientset
}

func NewClientset() *kubernetes.Clientset {
	var config *rest.Config
	var err error
	if kubeConfig == "incluster" {
		config, err = rest.InClusterConfig()
	} else {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	}
	if err != nil {
		log.Panicf("could not create k8s config: %s", err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Panicf("could not create k8s clientset: %s", err.Error())
	}

	return clientset
}

func NewKubeClient(l *logrus.Logger, c *kubernetes.Clientset) *KubeClient {
	log = l.WithFields(logrus.Fields{
		"component": "kubeclient",
	})

	k := &KubeClient{
		clientset: c,
	}

	return k
}

func (k *KubeClient) GetClientset() *kubernetes.Clientset {
	return k.clientset
}
