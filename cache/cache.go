// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package cache

import (
	"sync"
	"time"
)

var timeout = time.Minute

type Cache struct {
	Timeout time.Duration
	Mutex   sync.Mutex
	mutex   sync.Mutex
	data    map[string]*Item
}

type Item struct {
	data     interface{}
	deadline time.Time
}

func NewCache() *Cache {
	return &Cache{
		Timeout: timeout,
		data:    make(map[string]*Item),
	}
}

func (c *Cache) Set(key string, data interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = &Item{
		data:     data,
		deadline: time.Now().Add(c.Timeout),
	}
}

func (c *Cache) Get(key string) (data interface{}, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if i, ok := c.data[key]; ok {
		if c.Timeout.Seconds() == 0 || !time.Now().After(i.deadline) {
			return i.data, true
		}
	}
	return nil, false
}

func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.data[key]; ok {
		delete(c.data, key)
	}
}

func (c *Cache) Keys() []string {
	var a []string
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k := range c.data {
		a = append(a, k)
	}
	return a
}
