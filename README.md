

# mohawk

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## Introduction

Mohawk can use different [backends](/backend) for different use cases. Different backends may vary in speed, persistancy and scalability. Mohawk use a [RESTful API](/examples/REST.md) identical to Hawkular, inheriting Hawkular's echosystem of clients and plugins.

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

## Mock Certifications

The server requires certification to serve ``https`` requests. Users can use self signed credentials files for testing.

To create a self signed credentials use this bash commands:
```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

## Running the server

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

Running the server using the ``sqlite`` backend,
**Remmeber to set up the ``server.key`` and ``server.pem`` files in your path**.

```bash
mohawk -backend sqlite -tls -port 8443
2016/12/01 14:23:48 Start server, listen on https://0.0.0.0:8443
```

## Examples

For more examples look at the [example](/examples) directory.

### Running the TLS server on port 8443 supporting gzip encoding

Using TLS server requires certification files, default file names are `server.key` and `server.crt` .

```bash
mohawk -tls -gzip -port 8443
```

### Reading and writing data
```
# get server status
curl -ks  https://localhost:8443/hawkular/metrics/status

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

### Data encoding, using gzip data encoding

```
# using the zcat utility to decode gzip message
curl -k -s -H "Accept-Encoding: gzip" https://localhost:8443/hawkular/metrics/metrics | zcat
```
