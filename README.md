

# Mohawk

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Introduction

[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/mohawk)](https://goreportcard.com/report/github.com/yaacov/mohawk)

Mohawk can use different [backends](/backend) for different use cases. Different backends may vary in speed, persistancy and scalability. Mohawk use a subset of Hawkular's [REST API](/examples/REST.md), inheriting Hawkular's echosystem of clients and plugins.

Different use cases may have conflicting requirements for the metric engein, some use case may require fast data transfer, but no long term data retention, other may depend on long term, high availabilty data retention. We can use different metric data engins for each usecase, but then our consumer application will have to know how to interact with each different metric engien we choose.

Mohowk exposes the same simple REST API for different backend storage options, consumer application can use the same REST API with the fast, short term stroage and with the high availabilty, long term storage. Mohowk makes hierarchical data storage settings with short, middle and long term data retention easy to set up and consume.     

#### Plugins

|                  | Speed         |  Storage          | Advantages                                  |
|------------------|---------------|-------------------|---------------------------------------------|
| Memory           | Very Fast     | Memory            | No storage ware and tear from fast I/O      |
| Sqlite           | Fast          | Local File        | No data loss on network outages             | 
| Mongo            | Fast          | Mongo DB          | High availabilty, High volume storage       |

Mohawk architecture makes it easy to build and set up plugins for new data storage.

#### Banchmarks

1000 writes + 1000 reads, for more information see the [banchmark](/banchmark) directory.

| Backend  | Time       | %CPU      | RSS byte      |
|----------|------------|-----------|---------------|
|memory    |  0m2.011s  | 0.2 - 5.5 | 7456 - 11028  |
|mongo (*) |  0m4.885s  | 0.5 - 0.8 | 11892 - 11892 |
|sqlite3   |  0m14.471s | 0.2 - 7.4 | 8416 - 12560  |

(*) the mongo usage metrics does not include usage of the mongodb server.

#### Compatibility

Mohawk is tested(**) with [Hawkular](http://www.hawkular.org/) plugins, like [Hawkular Grafana Plugin](https://grafana.com/plugins/hawkular-datasource) and clients like [Python](https://github.com/hawkular/hawkular-client-python) and [Ruby](https://github.com/hawkular/hawkular-client-ruby)

(**) Mohawk implement only part of Hawkular's API, some functionalty may be missing.

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
