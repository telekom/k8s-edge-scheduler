// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package kubeclient

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KubeClient) GetLocationFromPod(p *v1.Pod) (string, error) {
	n, err := k.clientset.CoreV1().Nodes().Get(p.Spec.NodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return k.GetLocationFromNode(n)
}

func (k *KubeClient) GetLocationFromNode(n *v1.Node) (string, error) {
	if l, ok := n.Labels["location"]; ok {
		return l, nil
	}
	return "", fmt.Errorf("node %s has no label 'location'", n.Name)
}
