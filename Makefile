# k8s-edge-scheduler : custom kubernetes scheduler for placing pods based on location data
# Copyright (c) 2019, Lukas Steiner, Deutsche Telekom AG
# contact: opensource@telekom.de

# This file is licensed under the terms of the 3-Clause BSD License  [SPDX: BSD3-Clause]. 
# For Details see the file LICENSE on the top level of the project repository.

EDGE_SCHEDULER_IMAGE=k8s-edge-scheduler
EDGE_AGENT_IMAGE=k8s-edge-scheduler-agent
IMAGE_BUILD_FLAGS=-a -installsuffix cgo -ldflags '-extldflags \"-static\"'

build: build-scheduler build-agent
deploy: deploy-scheduler deploy-agent

clean:
	rm -r build

# scheduler
build-scheduler:
	go build -o build/edge-scheduler 

build-scheduler-linux:
	GOOS=linux CGO_ENABLED=0 go build -o build/edge-scheduler 

build-scheduler-image: build-scheduler-linux
	docker build -t $(EDGE_SCHEDULER_IMAGE) .

# agent
build-agent:
	go build -o build/edge-scheduler-agent ./agent

build-agent-linux:
	GOOS=linux CGO_ENABLED=0 go build -o build/edge-scheduler-agent ./agent

build-agent-image: build-agent-linux
	docker build -t $(EDGE_AGENT_IMAGE) -f agent/Dockerfile .
