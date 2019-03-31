// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package influxclient

import (
	client "github.com/influxdata/influxdb/client/v2"
	"github.com/namsral/flag"
)

var (
	influxAddr     string
	influxUser     string
	influxPassword string
	influxDB       string
)

func init() {
	flag.StringVar(&influxAddr, "influxAddr", "http://influxdb:8086", "influxdb address")
	flag.StringVar(&influxUser, "influxUser", "influx", "influxdb user")
	flag.StringVar(&influxPassword, "influxPassword", "influx", "influxdb user password")
	flag.StringVar(&influxDB, "influxDB", "edgescheduler", "influxdb database")
}

type InfluxClient struct {
	client client.Client
}

func NewInfluxClient() (*InfluxClient, error) {
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxAddr,
		Username: influxUser,
		Password: influxPassword,
	})
	if err != nil {
		return nil, err
	}

	return &InfluxClient{
		client: c,
	}, nil
}

func (i *InfluxClient) QueryDB(cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: influxDB,
	}
	if response, err := i.client.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func (i *InfluxClient) Close() {
	i.client.Close()
}
