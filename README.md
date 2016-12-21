# Flexswitch exporter [![Build Status](https://travis-ci.org/prometheus/flexswitch_exporter.svg)][travis]

[![CircleCI](https://circleci.com/gh/prometheus/flexswitch_exporter/tree/master.svg?style=shield)][circleci]
[![Docker Repository on Quay](https://quay.io/repository/prometheus/flexswitch-exporter/status)][quay]
[![Docker Pulls](https://img.shields.io/docker/pulls/prom/flexswitch-exporter.svg?maxAge=604800)][hub]

Prometheus exporter for SnapRoute FlexSwitch metrics, written in Go with pluggable metric
collectors.

## Collectors

You can create pluggable collectors for flexswitch similarly to that used by node_exporter.

### Enabled by default

Name     | Description
---------|-------------
ports | Shows FlexSwitch Port Utilization statistics 

## Building and running

    make
    ./flexswitch_exporter <flags>

## Running tests

    make test


## Using Docker

You can deploy this exporter using the [prom/flexswitch-exporter](https://registry.hub.docker.com/u/prom/flexswitch-exporter/) Docker image.

For example:

```bash
docker pull prom/flexswitch-exporter

docker run -d -p 9117:9117 --net="host" prom/flexswitch-exporter
```


[travis]: https://travis-ci.org/prometheus/flexswitch_exporter
[hub]: https://hub.docker.com/r/prom/flexswitch-exporter/
[circleci]: https://circleci.com/gh/prometheus/flexswitch_exporter
[quay]: https://quay.io/repository/prometheus/flexswitch-exporter
