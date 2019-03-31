// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package collector

import (
	"fmt"
	"regexp"
)

type deployment struct {
	name      string
	namespace string
}

func (c *Collector) getDeploymentNameByPodIP(ip string) (*deployment, error) {
	if i, ok := c.deploymentsByPodIP.Get(ip); ok {
		return i.(*deployment), nil
	}

	p, err := c.kube.GetPodByIPAllNamespaces(ip)
	if err != nil {
		return nil, err
	}

	if m, err := regexp.MatchString(nsIgnorePattern, p.Namespace); err == nil && m {
		c.blacklistedIPs.Set(ip, true)
		return nil, fmt.Errorf("blacklist ip %s, it's in %s namespace which matched ignore pattern", ip, p.Namespace)
	} else if err != nil {
		return nil, fmt.Errorf("namespace ignore pattern failed: %s", err.Error())
	}

	d, err := c.kube.GetDeploymentFromPod(p)
	if err != nil {
		return nil, err
	}

	i := &deployment{
		name:      d.Name,
		namespace: d.Namespace,
	}

	c.deploymentsByPodIP.Set(ip, i)

	return i, nil
}

func (c *Collector) getProxyLocation(name string) (string, error) {
	if i, ok := c.proxiesByName.Get(name); ok {
		return i.(string), nil

	}
	p, err := c.kube.GetPod(name, proxyNamespace)
	if err != nil {
		return "", err
	}

	l, err := c.kube.GetLocationFromPod(p)
	if err != nil {
		return "", err
	}

	c.proxiesByName.Set(name, l)

	return l, nil
}
