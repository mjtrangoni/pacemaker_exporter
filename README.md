# Pacemaker Exporter [![Build Status](https://travis-ci.org/mjtrangoni/pacemaker_exporter.svg)][travis]

[![CircleCI](https://circleci.com/gh/mjtrangoni/pacemaker_exporter.svg?style=svg)](https://circleci.com/gh/mjtrangoni/pacemaker_exporter)
[![GoDoc](https://godoc.org/github.com/mjtrangoni/pacemaker_exporter?status.svg)](https://godoc.org/github.com/mjtrangoni/pacemaker_exporter)
[![Coverage Status](https://coveralls.io/repos/github/mjtrangoni/pacemaker_exporter/badge.svg?branch=master)](https://coveralls.io/github/mjtrangoni/pacemaker_exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mjtrangoni/pacemaker_exporter)](https://goreportcard.com/report/github.com/mjtrangoni/pacemaker_exporter)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/00e03e600d5744d1a2cc21d98e2f8273)](https://www.codacy.com/app/mjtrangoni/pacemaker_exporter?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=mjtrangoni/pacemaker_exporter&amp;utm_campaign=Badge_Grade)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://raw.githubusercontent.com/mjtrangoni/pacemaker_exporter/master/LICENSE)

[Prometheus](https://prometheus.io/) exporter for [Pacemaker](https://github.com/ClusterLabs/pacemaker) cluster resource manager.

## Getting

```
$ go get github.com/mjtrangoni/pacemaker_exporter
```

## Building

```
$ cd $GOPATH/src/github.com/mjtrangoni/pacemaker_exporter
$ make
```

## Running

```
$ ./pacemaker_exporter <flags>
```
Note: Please run it as *root* user, otherwise `crm_mon` will be failing.
Alternatively, add user you run it as into haclient group.

## Endpoints

 1. http://localhost:9356/metrics for the Prometheus metrics.
 2. http://localhost:9356/html for a HTML cluster status page.
 2. http://localhost:9356/xml for a XML cluster status page.

## What's exported?

This exporter run `crm_mon -Xr`, and parse its XML output.

|   XML element    |     Status      | Default |
|:----------------:|:---------------:| :------:|
| summary          | implemented     | enabled |
| nodes            | implemented     | enabled |
| node_attributes  | implemented     | enabled |
| node_history     | not implemented |         |
| resources        | implemented     | enabled |
| resources/bundle | not implemented |         |
| resources/group  | implemented     | enabled |
| resources/clone  | implemented     | enabled |
| tickets          | not implemented |         |
| bans             | implemented     | enabled |
| failures         | implemented     | enabled |

## Dashboards

 1. [TODO:Grafana Dashboard]()

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/mjtrangoni/pacemaker_exporter/blob/master/CONTRIBUTING.md)

## License

Apache License 2.0, see [LICENSE](https://github.com/mjtrangoni/mjtrangoni/blob/master/LICENSE).

[travis]: https://travis-ci.org/mjtrangoni/pacemaker_exporter
