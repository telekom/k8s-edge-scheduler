# k8s-edge-scheduler

A kubernetes scheduler, which places pods based on tracing data from reverse proxies to reduce latency.

## Build

    make build

Artifacts can be found in `./build`

##  Deploy

Kubernetes resource examples are placed in `./deploy`.
An influx database is required.
Kubernetes nodes need a label named `location` with a city or region as value, 
so that the scheduler can locate the reverse proxies and applications
