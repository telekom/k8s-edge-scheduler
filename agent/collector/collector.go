// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package collector

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
	"github.com/telekom/k8s-edge-scheduler/cache"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

const writePeriod = 15 * time.Second

var (
	log             *logrus.Entry
	influxAddr      string
	influxUser      string
	influxPassword  string
	influxDB        string
	proxyNamespace  string
	nsIgnorePattern string
	databasePrefix  string
)

func init() {
	flag.StringVar(&influxAddr, "influxAddr", "http://influxdb:8086", "influxdb address")
	flag.StringVar(&influxUser, "influxUser", "influx", "influxdb user")
	flag.StringVar(&influxPassword, "influxPassword", "influx", "influxdb user password")
	flag.StringVar(&proxyNamespace, "proxyNamespace", "open-edge-cloud", "proxy namespace")
	flag.StringVar(&nsIgnorePattern, "nsIgnoreRegex", "", "ignore pattern for system namespaces")
	flag.StringVar(&databasePrefix, "databasePrefix", "edge-", "database prefix")
}

type Collector struct {
	influx client.Client
	mutex  sync.Mutex
	batch  map[string]client.BatchPoints
	kube   KubernetesClient

	deploymentsByPodIP *cache.Cache
	proxiesByName      *cache.Cache
	blacklistedIPs     *cache.Cache
}

type KubernetesClient interface {
	GetPodByIPAllNamespaces(ip string) (*v1.Pod, error)
	GetPod(name string, namespace string) (*v1.Pod, error)
	GetDeploymentFromPod(p *v1.Pod) (*appsv1.Deployment, error)
	GetLocationFromPod(p *v1.Pod) (string, error)
}

func NewCollector(k KubernetesClient, l *logrus.Logger) *Collector {
	log = l.WithFields(logrus.Fields{
		"component": "collector",
	})

	i, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxAddr,
		Username: influxUser,
		Password: influxPassword,
	})
	if err != nil {
		log.Fatalf("cannot connect to influxdb: %s", err.Error())
	}
	log.Infof("connected to influxdb %s", influxAddr)

	c := &Collector{
		influx:             i,
		batch:              make(map[string]client.BatchPoints),
		deploymentsByPodIP: cache.NewCache(),
		proxiesByName:      cache.NewCache(),
		blacklistedIPs:     cache.NewCache(),
		kube:               k,
	}
	c.blacklistedIPs.Timeout = time.Hour

	return c
}

func (c *Collector) Start() {
	defer c.influx.Close()
	for {
		c.write()
		<-time.NewTimer(writePeriod).C
	}
}

func (c *Collector) AddRequest(ip string, proxy string, hostname string, timestamp int64, duration int64) {
	if _, ok := c.blacklistedIPs.Get(ip); ok {
		log.Debugf("skipping blacklisted ip %s", ip)
		return
	}

	d, err := c.getDeploymentNameByPodIP(ip)
	if err != nil {
		log.Warn(err.Error())
		return
	}
	src, err := c.getProxyLocation(hostname)
	if err != nil {
		log.Warn(err.Error())
		return
	}

	db := fmt.Sprintf("%s%s", databasePrefix, d.namespace)

	log.Debugf("ip: %s, proxy: %s, timestamp: %d, source: %s, destination: %s, database: %s", ip, proxy, timestamp, src, d.name, db)

	tags := map[string]string{
		"location": src,
		"app":      d.name,
	}
	fields := map[string]interface{}{
		"duration":    duration,
		"proxy":       proxy,
		"source":      src,
		"destination": d.name,
	}

	t := strconv.FormatInt(timestamp, 10)
	sep := len(t) - 6
	sec, _ := strconv.ParseInt(t[:sep], 10, 64)
	usec, _ := strconv.ParseInt(t[sep:], 10, 64)
	pt, err := client.NewPoint("request", tags, fields, time.Unix(sec, usec*1000))
	if err != nil {
		log.Warnf("cannot create new point: %s", err.Error())
	} else if err := c.add(pt, db); err != nil {
		log.Warnf("cannot add new point: %s", err.Error())
	}
}

func (c *Collector) add(p *client.Point, db string) error {
	var err error
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if _, ok := c.batch[db]; !ok {
		c.batch[db], err = client.NewBatchPoints(client.BatchPointsConfig{
			Database:  db,
			Precision: "ms",
		})
		if err != nil {
			return err
		}
	}
	c.batch[db].AddPoint(p)
	return nil
}

func (c *Collector) write() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, p := range c.batch {
		l := len(p.Points())
		if l > 0 {
			if err := c.influx.Write(p); err != nil {
				log.Warnf("cannot write points do database %s: %s", p.Database(), err.Error())
			} else {
				log.Debugf("wrote %d points to database %s", l, p.Database())
			}
		}
	}
	c.batch = make(map[string]client.BatchPoints)
}
