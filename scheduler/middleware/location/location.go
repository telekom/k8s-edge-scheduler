// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package location

import (
	"fmt"
	"strings"

	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware"
	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware/location/influxclient"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	name = "location"
)

var (
	defaultLocation string
	namespace       string
	log             *logrus.Entry
	timeRanges      = []struct {
		time  string
		multi int
	}{
		{time: "15m", multi: 3},
		{time: "1h", multi: 2},
		{time: "24h", multi: 1},
	}
)

func init() {
	flag.StringVar(&defaultLocation, "defaultLocation", "", "default location")
}

func Location(m middleware.Middleware) middleware.Middleware {
	return func(s middleware.Scheduler, d *middleware.Data) {
		log = s.Log(name)
		defer m(s, d)

		s.GetNodes().Mutex.Lock()
		for _, k := range d.Prio.Keys() {
			o, _ := s.GetNodes().Get(k)
			n := o.(*v1.Node)
			l, err := s.GetKube().GetLocationFromNode(n)
			if err != nil {
				log.Warn(err.Error())
				continue
			}

			// best location
			if p := getLocationPoints(d.Deployment, l); p != 0 {
				log.Debugf("node %s gets %d points for placed at location %s", n.Name, p, l)
				if err := d.Prio.Add(k, p); err != nil {
					log.Warn(err.Error())
				}
			}

			// default location
			if defaultLocation != "" && l == defaultLocation {
				log.Debugf("node %s gets 5 points for placed at default location %s", n.Name, l)
				if err := d.Prio.Add(k, 5); err != nil {
					log.Warn(err.Error())
				}
			}

			// location tolerances
			if !isTolerated(d.Pod, l) {
				d.Prio.Disable(k)
				log.Debugf("deny scheduling pod %s to node %s, because of location tolerances", d.Pod.Name, k)
			}
		}
		s.GetNodes().Mutex.Unlock()
	}
}

func getLocationPoints(d *appsv1.Deployment, location string) int {
	i, err := influxclient.NewInfluxClient()
	if err != nil {
		log.Warn(err.Error())
		return 0
	}
	defer i.Close()

	for _, r := range timeRanges {
		p, err := getLocationRequestPercent(i, d, r.time, location)
		if err != nil {
			log.Warn(err.Error())
			break
		} else if p != 0 {
			return (p / 10) * r.multi
		}
	}
	return 0
}

func isTolerated(p *v1.Pod, location string) bool {
	if d, err := isInLabel("deniedLocations", p, location); err == nil {
		return !d
	} else if d, err := isInLabel("allowedLocations", p, location); err == nil {
		return d
	}
	return true
}

func isInLabel(label string, p *v1.Pod, key string) (bool, error) {
	if v, ok := p.Labels[label]; !ok {
		return false, fmt.Errorf("label %s not set", label)
	} else {
		for _, k := range strings.Split(v, ",") {
			if k == key {
				return true, nil
			}
		}
	}
	return false, nil
}
