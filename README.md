

# mohawk

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")


MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## Introduction

[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/mohawk)](https://goreportcard.com/report/github.com/yaacov/mohawk)

Mohawk can use different [backends](/backend) for different use cases. Different backends may vary in speed, persistancy and scalability. Mohawk use a subset of Hawkular's [RESTful API](/examples/REST.md), inheriting Hawkular's echosystem of clients and plugins.

#### Plugins

| Backend |                                                                                                                                      |
|---------|--------------------------------------------------------------------------------------------------------------------------------------|
| Example | Backend template. Dump random data on data requests, fails silently on unimplemented requests.                                       |
| Sqlite  | File storage based backend. Fast, low on system resources, use local files for data persistency.                                     |
| Memory  | Memory storage based backend. Very fast, very low on system resources, 7 day data Retention, no data persistency after process end.  |

#### Compatibility - Mohawk is tested with Hawkular plugins, like [Hawkular Grafana Plugin](https://grafana.com/plugins/hawkular-datasource) and clients like [Python](https://github.com/hawkular/hawkular-client-python) and [Ruby](https://github.com/hawkular/hawkular-client-ruby)

## Installation

Using a Copr repository for Fedora:

```
sudo dnf copr enable yaacov/mohawk
sudo dnf install mohawk
```

Using Dockerhub repository:

```
docker run -v [PATH TO KEY AND CERT FILES]:/root/ssh:Z yaacov/mohawk
```

## Running the server

#### Mock Certifications

The server requires certification to serve ``https`` requests. Users can use self signed credentials files for testing.

To create a self signed credentials use this bash commands:
```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

#### Running the server

Request usage message.

```bash
mohawk -h
Usage of mohawk:
...
```

Running ``mohawk`` without ``tls`` and using the ``sqlite`` back end.

```bash
mohawk
2017/01/03 10:06:50 Start server, listen on http://0.0.0.0:8080
```

Running the server with ``tls``, ``gzip`` encoding support and using the ``memory`` backend,
**Remmeber to set up the ``server.key`` and ``server.pem`` files in your path**.

```bash
mohawk -backend memory -tls -gzip -port 8443
2016/12/01 14:23:48 Start server, listen on https://0.0.0.0:8443
```

## Examples

For more in-depth usage information look at the [example](/examples) directory.

#### Running the server for this examples

Using TLS server requires certification files, default file names are `server.key` and `server.pem` .

```bash
mohawk -tls -gzip -port 8443
```

#### Reading and writing data
```
# get server status
curl -ks https://localhost:8443/hawkular/metrics/status

# get a list of all metrics
curl -ks https://localhost:8443/hawkular/metrics/metrics

# post some data (timestamp is in ms)
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d "[{\"id\":\"machine/example.com/test\", \"data\":[{\"timestamp\": 1492434911769, \"value\": 42}]}]"

# read some data (variables can be start, end and bucketDuration)
curl -ks https://localhost:8443/hawkular/metrics/gauges/machine%2Fexample.com%2Ftest/raw?start=1492434911760

# set tags
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/machine%2Fexample.com%2Ftest/tags -d "{\"type\": \"node\", \"hostname\": \"example.com\"}"

# look for metrics by tag value (using a regexp)
curl -ks https://localhost:8443/hawkular/metrics/metrics?tags=hostname:.*\.com

# read multiple data points
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d "{\"ids\": [\"machine/example.com/test\"], \"start\": 1492434811769, \"end\": 1492435911769}"

# read multiple data points with aggregation statistics
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d "{\"ids\": [\"machine/example.com/test\"], \"start\": 1492434811769, \"end\": 1492435911769, \"bucketDuration\": \"1000s\"}"
```

#### Data encoding, using gzip data encoding

```
# using the zcat utility to decode gzip message
curl -ks -H "Accept-Encoding: gzip" https://localhost:8443/hawkular/metrics/metrics | zcat
```
