// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package kubeclient

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *KubeClient) GetPodByIP(ip string, namespace string) (*v1.Pod, error) {
	pods, err := k.clientset.CoreV1().Pods(namespace).List(metav1.ListOptions{
		FieldSelector: "status.podIP=" + ip,
	})
	if err != nil {
		log.Warn(err.Error())
		return nil, err
	}

	c := len(pods.Items)
	if c == 0 {
		return nil, fmt.Errorf("no pod with ip %s found", ip)
	} else if c > 1 {
		err := fmt.Errorf("multiple pods with ip %s found", ip)
		log.Warn(err.Error())
		return nil, err
	}

	return &pods.Items[0], nil
}

func (k *KubeClient) GetPod(name string, namespace string) (*v1.Pod, error) {
	pod, err := k.clientset.CoreV1().Pods(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (k *KubeClient) GetPodByIPAllNamespaces(ip string) (*v1.Pod, error) {
	namespaces, err := k.clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, n := range namespaces.Items {
		pod, err := k.GetPodByIP(ip, n.Name)
		if err == nil {
			return pod, nil
		}
	}

	return nil, fmt.Errorf("no pod with ip %s found in any namespace", ip)
}

func (k *KubeClient) GetDeploymentFromPod(p *v1.Pod) (*appsv1.Deployment, error) {
	if len(p.OwnerReferences) == 0 {
		return nil, fmt.Errorf("pod %s has no owner reference", p.Name)
	}
	r, err := k.clientset.AppsV1().ReplicaSets(p.Namespace).Get(p.OwnerReferences[0].Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("replica set %s not found", p.OwnerReferences[0].Name)
	}
	if len(r.OwnerReferences) == 0 {
		return nil, fmt.Errorf("replicaset %s has no owner reference", r.Name)
	}
	d, err := k.clientset.AppsV1().Deployments(r.Namespace).Get(r.OwnerReferences[0].Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("deployment %s not found", r.OwnerReferences[0].Name)
	}
	return d, nil
}

func (k *KubeClient) GetPodsFromDeployment(d *appsv1.Deployment) ([]v1.Pod, error) {
	s, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
	if err != nil {
		return nil, err
	}

	rs, err := k.clientset.AppsV1().ReplicaSets(d.Namespace).List(metav1.ListOptions{
		LabelSelector: s.String(),
	})
	if err != nil {
		return nil, err
	}

	if len(rs.Items) == 0 {
		return nil, fmt.Errorf("deployment %s has no replicaset", d.Name)
	}
	var l []v1.Pod
	for _, i := range rs.Items {
		p, err := k.GetPodsFromReplicaSet(&i)
		if err != nil {
			log.Warn(err.Error())
		} else {
			l = append(l, p...)
		}
	}
	return l, nil
}

func (k *KubeClient) GetPodsFromReplicaSet(r *appsv1.ReplicaSet) ([]v1.Pod, error) {
	s, err := metav1.LabelSelectorAsSelector(r.Spec.Selector)
	if err != nil {
		return nil, err
	}

	p, err := k.clientset.CoreV1().Pods(r.Namespace).List(metav1.ListOptions{
		LabelSelector: s.String(),
	})
	if err != nil {
		return nil, err
	}

	return p.Items, nil
}
