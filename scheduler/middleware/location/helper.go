// k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
// Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
// contact: opensource@telekom.de

// This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause].
// For Details see the file LICENSE on the top level of the project repository.

package location

import (
	"encoding/json"
	"fmt"

	"github.com/telekom/k8s-edge-scheduler/scheduler/middleware/location/influxclient"
	appsv1 "k8s.io/api/apps/v1"
)

func getLocationRequestPercent(i *influxclient.InfluxClient, d *appsv1.Deployment, timeRange string, location string) (int, error) {
	r, err := i.QueryDB(fmt.Sprintf("SELECT count(\"duration\") FROM \"request\" WHERE (\"app\" = '%s') AND (\"location\" = '%s') AND time >= now() - %s", d.Name, location, timeRange))
	if err != nil {
		return 0, err
	}

	if len(r) == 0 || len(r[0].Series) == 0 {
		return 0, nil
	}
	c, err := r[0].Series[0].Values[0][1].(json.Number).Int64()
	if err != nil {
		return 0, err
	}
	// log.Debugf("deployment %s called %d times from location %s in timerange %s", d.Name, c, location, timeRange)

	r, err = i.QueryDB(fmt.Sprintf("SELECT count(\"duration\") FROM \"request\" WHERE (\"app\" = '%s') AND time >= now() - %s", d.Name, timeRange))
	if err != nil {
		return 0, err
	}

	if len(r) == 0 || len(r[0].Series) == 0 {
		return 0, fmt.Errorf("something is wrong with the database")
	}

	a, err := r[0].Series[0].Values[0][1].(json.Number).Int64()
	if err != nil {
		return 0, err
	}
	// log.Debugf("deployment %s called %d times from all locations in timerange %s", d.Name, a, timeRange)

	return int((c * 100) / a), nil
}
