

# Mohawk

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Introduction

[![Go Report Card](https://goreportcard.com/badge/github.com/yaacov/mohawk)](https://goreportcard.com/report/github.com/yaacov/mohawk)

Mohawk can use different storage [plugins](/backend) for different use cases. Different storage plugins may vary in speed, persistancy and scalability. Mohawk use a subset of Hawkular's [REST API](/examples/REST.md), inheriting Hawkular's echosystem of clients and plugins.

Different use cases may have conflicting requirements for the metric engein, some use cases may require fast data transfer, while others may depend on long term, high availabilty data retention that inherently makes the system slower.

Mohowk exposes the same simple REST API for different backend storage options, consumer application can use the same REST API with a lean low footprint stroage and with a resource-intensive high availabilty storage. Mohowk makes hierarchical data storage using short, middle and long term data retention tiers easy to set up and consume.     

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

#### Storage Plugins

Mohawk architecture makes it easy to implement and set up [plugins](/backend) for new data storage. The backend directory include documentation, examples and a template for plugin develpment.

###### Current storage plugin list include:

| Plugin name       |  Storage          | Advantages                                  | Use case                                 |
|-------------------|-------------------|---------------------------------------------|------------------------------------------|
| memory            | Memory            | No storage ware and tear from fast I/O      | Fast I/O, no need for persistence data   |
| sqlite            | Local File        | No data loss on network outages             | Persistence data, W/O external data base |
| mongo             | Mongo DB          | High availabilty, High volume storage       | Long term H.A. storage                   |

A template plugin named [example](/backend/example) is also available.

#### Benchmarks

###### Name: Run time, real
###### Description: 1000 writes + 1000 reads ( [benchmark.py](/benchmark/benchmark.py) ) less is better.

Benchmark results depend on system resurcses, current work load and network.
For more information on tests and run enviroments, see the [benchmark](/benchmark) directory.

###### Mohawk with different Plugins running on a desktop machine.

| Plugin   | Time       | %CPU      | RSS byte      |
|----------|------------|-----------|---------------|
|memory    |  0m2.011s  | 0.2 - 5.5 | 7456 - 11028  |
|mongo (1) |  0m4.885s  | 0.5 - 0.8 | 11892 - 11892 |
|sqlite3   |  0m14.471s | 0.2 - 7.4 | 8416 - 12560  |

(1) the mongo usage metrics does not include usage of the mongodb server.

###### Chart: different Plugins vs. Run Time

![Time chart](/benchmark/time.png?raw=true "benchmark time vm")

###### Mohawk vs. Hawkular running on a vm under same load.

| DB/Plugin          | Time        |
|---------------------|-------------|
|Hawkular/Casandra    |  2m8.783s   |
|Mohawk/Memory        |  0m22.833s  |

###### Chart: DB/Plugin vs. Run Time

![Time chart](/benchmark/time-vm.png?raw=true "benchmark time vm")

#### Performance

Moahawk cpu and memory usage is lower than Hawkular and comparable to Prometheus, for more details see [Performance](/benchmark/PERF.md) doc.

###### Mohawk vs. Prometheus CPU (Pod name is hawkular-metrics, but actually running mohawk)

![CPU chart](/benchmark/mohawk-cpu.png?raw=true "benchmark cpu vm")
![CPU chart](/benchmark/prometheus-cpu.png?raw=true "benchmark cpu vm")

###### Mohawk vs. Prometheus Memory (Pod name is hawkular-metrics, but actually running mohawk)

![CPU chart](/benchmark/mohawk-mem.png?raw=true "benchmark cpu vm")
![CPU chart](/benchmark/prometheus-mem.png?raw=true "benchmark cpu vm")

#### Compatibility

Mohawk is tested(2) with [Hawkular](http://www.hawkular.org/) plugins, like [Hawkular Grafana Plugin](https://grafana.com/plugins/hawkular-datasource) and clients like [Python](https://github.com/hawkular/hawkular-client-python) and [Ruby](https://github.com/hawkular/hawkular-client-ruby). Mohawk also work with [Heapster](https://github.com/kubernetes/heapster).

(2) Mohawk implement only part of Hawkular's API, some functionalty may be missing.

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

#### Syndication and hierarchical server setup

It is easy to scrape data from one server to another, creating a hierarchy of servers
with a central server collecting specific data from the peripheral servers.

The [syndication.sh](/examples/syndication.sh) script is an example of a script that scrape
data from one server to another. Scraping scripts can be simple or complicated as needed.

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
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/tags -d "[{\"id\":\"machine/example.com/test\", \"tags\":{\"type\": \"node\", \"hostname\": \"example.com\"}}]"

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

```
# sendig gziped data file with curl's --data-binary flag
curl -ks -H "Content-Encoding: gzip" -X PUT "https://localhost:8443/hawkular/metrics/gauges/tags" --data-binary @tags.json.gz
```
