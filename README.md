[![Go Report Card](https://goreportcard.com/badge/github.com/MohawkTSDB/mohawk)](https://goreportcard.com/report/github.com/MohawkTSDB/mohawk)
[![Build Status](https://travis-ci.org/MohawkTSDB/mohawk.svg?branch=master)](https://travis-ci.org/MohawkTSDB/mohawk)

# Mohawk

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine, it's fun, fast, light and easy to use.

#### Table of contents

  - [Introduction](#introduction)
  - [Installation](#installation)
  - [Running the server](#running-the-server)
  - [Reading and writing data](#reading-and-writing-data)

#### Source Modules

  - [REST API (server)](/src/server/) source directory
  - [Command Line Interface (cli)](/src/cli/) source directory
  - [Metric Data Storage (storage)](/src/storage/) source directory
  - [Alert Rules (alerts)](/src/alerts/) source directory

## Introduction

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

Mohawk can use different storage plugins for different use cases. Different storage plugins may vary in speed, persistence and scale ability. Mohawk use a subset of Hawkular's REST API inheriting Hawkular's ecosystem of clients and plugins.

#### Compatibility

Mohawk is tested(1) with [Hawkular](http://www.hawkular.org/) plugins, like [Hawkular Grafana Plugin](https://grafana.com/plugins/hawkular-datasource) and clients like [Python](https://github.com/hawkular/hawkular-client-python) and [Ruby](https://github.com/hawkular/hawkular-client-ruby). Mohawk also work with [Heapster](https://github.com/kubernetes/heapster) to automagically scrape metrics from [Kubernetes](https://kubernetes.io/) / [OpenShift](https://www.openshift.com/) clusters.

Mohawk can also serve as [Prometheus](https://prometheus.io/) scraping endpoint.

(1) Mohawk implement only a subset of Hawkular's API, some functionality may be missing.

## Installation

#### Using a Copr repository for Fedora

```
sudo dnf copr enable yaacov/mohawk
sudo dnf install mohawk
```

#### Using Dockerhub repository

```
docker run -v [PATH TO KEY AND CERT FILES]:/root/ssh:Z yaacov/mohawk
```

#### Building from source

```
# Create a directory for sources
mkdir -p ${GOPATH}/src/github.com/MohawkTSDB && cd ${GOPATH}/src/github.com/MohawkTSDB

# Clone the sources from the git repository
git clone https://github.com/MohawkTSDB/mohawk.git
cd mohawk

# Update vedor sources
make vendor

# Build, test and install
make clean
make
make test
make install

# Set up tls secrets for testing
make secret
```

## Running the server

#### The server command line interface

Request usage message.

```bash
mohawk -h
Usage of mohawk:
...
```

```bash
mohawk --options=help
Storage options:
...
```

Running ``mohawk`` without ``tls`` and using the ``sqlite`` back end.

```bash
mohawk --storage sqlite
2017/01/03 10:06:50 Start server, listen on http://0.0.0.0:8080
```

Running the server with ``tls``, ``gzip`` encoding support and using the ``memory`` storage.

###### Remember to set up the ``server.key`` and ``server.pem`` files in your path

```bash
mohawk --storage memory --tls --gzip --port 8443
2016/12/01 14:23:48 Start server, listen on https://0.0.0.0:8443
```
###### Mock Certifications for testing

The server requires certification to serve ``https`` requests. Users can use self signed credentials files for testing.

To create a self signed credentials use this bash commands:
```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```
If running from source, the Makefile has a utility for generating secrets:
```
make secret
```

#### Running the server for this examples

Using TLS server requires certification files, default file names are `server.key` and `server.pem` .

```bash
mohawk --tls --gzip --port 8443
```

## Reading and writing data

#### Common queries

```
# get server status
curl -ks https://localhost:8443/hawkular/metrics/status

# get a list of all metrics
curl -ks https://localhost:8443/hawkular/metrics/metrics

# post some data (timestamp is in ms)
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw \
     -d "[{\"id\":\"machine/example.com/test\", \"data\":[{\"timestamp\": 1492434911769, \"value\": 42}]}]"

# read some data (variables can be start, end and bucketDuration)
curl -ks https://localhost:8443/hawkular/metrics/gauges/machine%2Fexample.com%2Ftest/raw?start=1492434911760
```
#### Tags
```
# set tags
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/tags \
     -d "[{\"id\":\"machine/example.com/test\",\"tags\":{\"type\":\"node\",\"hostname\":\"example.com\"}}]"

# look for metrics by tag value (using a regexp)
curl -ks https://localhost:8443/hawkular/metrics/metrics?tags=hostname:.*\.com
```
#### Search
```
# read multiple data points using an ids list
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query \
     -d "{\"ids\": [\"machine/example.com/test\"], \"start\": 1492434811769, \"end\": 1492435911769}"

# read multiple data points using a regexp search on tags
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query \
     -d "{\"tags\": \"name:.*test|.*machine\", \"start\": \"-12h\"}"
```
#### Aggregation
```
# using relative start time, and aggregation time of 2mn
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query \
     -d "{\"ids\": [\"machine/example.com/test\"], \"start\":  \"-6h\", \"bucketDuration\": \"2mn\"}"
```

#### Data encoding, using gzip data encoding

```
# using the zcat utility to decode gzip message
curl -ks -H "Accept-Encoding: gzip" https://localhost:8443/hawkular/metrics/metrics | zcat
```

```
# sendig gziped data file with curl's --data-binary flag
curl -ks -H "Content-Encoding: gzip" -X PUT "https://localhost:8443/hawkular/metrics/gauges/tags" \
     --data-binary @tags.json.gz
```


#### Issues? Questions?

[Github Project](https://github.com/MohawkTSDB) / [Mohawk TSDB](https://mohawktsdb.github.io)
