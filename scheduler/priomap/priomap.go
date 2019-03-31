// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package priomap

import (
	"fmt"
	"sort"
	"sync"

	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
)

var (
	max = 100
	min = 0
)

type NodePrioMap struct {
	mutex sync.Mutex
	data  map[string]int
}

func NewNodePrioMap(nodes []string) *NodePrioMap {
	prio := make(map[string]int)
	for _, n := range nodes {
		prio[n] = 20
	}

	return &NodePrioMap{
		data: prio,
	}
}
func (n *NodePrioMap) Get(node string) (int, error) {
	if p, ok := n.data[node]; ok {
		return p, nil
	}
	return 0, fmt.Errorf("node %s not in map", node)
}

func (n *NodePrioMap) Set(node string, v int) error {
	if v > max || v < 0 {
		return fmt.Errorf("can't set to more than %d or less than null points", max)
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.data[node] = v
	return nil
}

func (n *NodePrioMap) Add(node string, add int) error {
	if add > max || add < -max {
		return fmt.Errorf("can't add or reduce more than %d points", max)
	}
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if p, ok := n.data[node]; ok {
		if p != -1 {
			if np := p + add; np < min {
				n.data[node] = min
			} else if np > max {
				n.data[node] = max
			} else {
				n.data[node] = np
			}
			return nil
		}
		return fmt.Errorf("node %s is already disabled", node)
	}
	return fmt.Errorf("node %s not in map", node)
}

func (n *NodePrioMap) Disable(node string) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	if _, ok := n.data[node]; ok {
		n.data[node] = -1
		return nil
	}
	return fmt.Errorf("node %s not in map", node)
}

func (n *NodePrioMap) Map() map[string]int {
	return n.data
}

func (n *NodePrioMap) Max() string {
	var node string
	max := -1
	n.mutex.Lock()
	defer n.mutex.Unlock()
	for n, p := range n.data {
		if p > max {
			node = n
			max = p
		}
	}
	return node
}

func (n *NodePrioMap) ListDec() []middleware.PrioMapPair {
	d := n.List()
	sort.Slice(d, func(i, j int) bool {
		return d[i].Value > d[j].Value
	})
	return d
}

func (n *NodePrioMap) ListInc() []middleware.PrioMapPair {
	d := n.List()
	sort.Slice(d, func(i, j int) bool {
		return d[i].Value < d[j].Value
	})
	return d
}

func (n *NodePrioMap) List() []middleware.PrioMapPair {
	var d []middleware.PrioMapPair
	n.mutex.Lock()
	defer n.mutex.Unlock()
	for k, v := range n.data {
		d = append(d, middleware.PrioMapPair{
			Key:   k,
			Value: v,
		})
	}
	return d
}

func (n *NodePrioMap) Keys() []string {
	var keys []string
	n.mutex.Lock()
	defer n.mutex.Unlock()
	for k := range n.data {
		keys = append(keys, k)
	}
	return keys
}
