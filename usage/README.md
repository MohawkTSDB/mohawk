

# mohawk/usage

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Usage

When installed, run using the command line ``mohawk``

```bash
mohawk --version

Mohawk version: 0.22.1
```

The `-h` flag will print out a help text, that list the command line arguments.

```
mohawk -h

NAME:
   mohawk - Metric data storage engine

USAGE:
   mohawk [global options] command [command options] [arguments...]

VERSION:
   0.18.1

AUTHOR:
   Yaacov Zamir <kobi.zamir@gmail.com>

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --storage value, -b value  the storage plugin to use (default: "memory")
   --token value              authorization token
   --key value                path to TLS key file (default: "server.key")
   --cert value               path to TLS cert file (default: "server.pem")
   --options value            specific storage options [e.g. db-dirname, db-url]
   --port value, -p value     server port (default: 8080)
   --tls, -t                  use TLS server
   --gzip, -g                 enable gzip encoding
   --verbose, -V              more debug output
   --help, -h                 show help
   --version, -v              print the version
```

Running ``mohawk`` with ``tls`` and using the ``memory`` back end.

```
mohawk --tls --gzip --port 8443 --storage memory
2017/06/30 11:37:08 Start server, listen on https://0.0.0.0:8443
```

###### Examples below use this server configuration, since each storage may implement different feature set, responses may be a little different for different plugins.

###### Running with tls on, we need .key and .pem files:

```
openssl ecparam -genkey -name secp384r1 -out server.key
openssl req -new -x509 -sha256 -key server.key -out server.pem -days 3650
```

#### JSON + REST API
JSON over [REST API](/usage/REST.md) is the primary interface of Mohawk Metrics. This makes it easier for users to get started and also makes integration easier since REST+JSON is widely used and easily understood. a rich, growing set of features that includes:

#### Multi Tenancy
Mohawk Metrics provides virtual multi tenancy. All data is mapped to a tenant. Everything is partitioned by tenant. All requests, both reads and writes, can include a tenant id, default tenant id is `_ops`.

#### Tagging
Mohawk Metrics provides flexible tagging support that makes it easy to organize and group data. Tagging can also be used to provide additional information and context about data.

#### Querying
Mohawk Metrics offers a rich set of features around querying that are ideal for rendering data in graphs and in charts. This includes:

  - Filtering and grouping with tags
  - Searching metric definitions
  - Downsampling and aggregation
  - Limit and order results

#### Token authentication

If the token flag is set, all API requests must send the valid bearer token in the "Authorization" request header.

```
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d @data.json -H "Authorization: Bearer THE_TOKEN" \
```

#### Tenants

All data is partitioned by tenant. The partitioning happens at the API level. This means that a metric cannot exist on its own outside of a tenant.

#### Implicit tenant creation

```
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d @data.json -H "Hawkular-Tenant: com.acme"
```

This is a request to insert gauge data points for the com.acme tenant. If that tenant does not already exist, it will be request when storing the metric data. Specific details on inserting data can be found in Inserting Data.

#### Tenant Header

As previously stated all data is partitioned by tenant. Mohawk Metrics enforces this by allowing the Hawkular-Tenant HTTP header in requests. The value of the header is the tenant id. We saw this already with the implicit tenant creation.

Using the Hawkular-Tenant HTTP header in request:

```
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics -H "Hawkular-Tenant: com.acme"
```

###### All requests, both reads and writes, can include a tenant id, default tenant id is `_ops`. If no tenant id is provided the `_ops` tenant will be used.

#### Tenant Ids

A tenant has an id that uniquely identifies it. The id is a variable length, UTF-8 encoded string. Mohawk Metrics does not perform any validation checks to prevent duplicate ids. If the key already exists in the map, it will simply be overwritten with the new value.

#### Inserting Data

Inserting data is a synchronous operation with respect to the client. An HTTP response is not returned until all data points are inserted. On the server side however, multiple inserts to the database are done in parallel to achieve higher throughput.

#### Data Points

A data point in Mohawk Metrics is a tuple that in its simplest form consists of a timestamp and a value.

##### Timestamps

Timestamps are unix timestamps in milliseconds.

##### Insert data points

```
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d @data.json
```

data.json

```json
[
  {
    "id": "free_memory",
    "data": [
      {"timestamp": 1460111065369, "value": 2048},
      {"timestamp": 1460151065352, "value": 2012},


      {"timestamp": 1460711012361, "value": 2012}
    ]
  },
  {
    "id": "cpu_usage",
    "data": [
      {"timestamp": 1460111065369, "value": 1.34},
      {"timestamp": 1460151085344, "value": 0.45},


      {"timestamp": 1460711075351, "value": 1.34}
    ]
  }
]
```

Each array element is an object that has id and data properties. data contains an array of data points.

#### Tagging

Tags in Mohawk Metrics are key/value pairs. Tags can be applied to a metric to provide meta data for the time series as a whole. Tags can be used to perform filtering in queries.

#### Updating Metric Tags

These endpoints are used to add or replace tags.

```
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/tags -d @tags.json
```

tags.json

```json
[
  {
    "id": "free_memory",
    "tags": {
      "datacenter": "dc2",
      "hostname": "server1"
    }
  }
]
```

#### Tag Filtering

Mohawk Metrics provides regular expression support for tag value filtering.

| Type           | Example       |                                                               |
|----------------|---------------|---------------------------------------------------------------|
| tag_name:regex | hostname:.*01 | Search for tag named hostname with a value that ends with 01. |

#### Querying

The examples provided in the following sections are not an exhaustive listing of the full API.

#### These operations do not fetch data points but rather the metric definition itself.

The next example illustrates the type parameter which filters the results by the specified types.

Fetch all metric definitions

```
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics
```

response body

```json
[
  {
    "id": "free_memory",
    "type": "gauge",
    "tags": {
      "datacenter": "dc2",
      "hostname": "server1"
    },
    "data": [
      {
        "timestamp": 1460711012361,
        "value": 2012
      }
    ]
  },
  {
    "id": "cpu_usage",
    "type": "gauge",
    "tags": {},
    "data": [
      {
        "timestamp": 1460711075351,
        "value": 1.34
      }
    ]
  }
]
```

The next example demonstrates querying metric and filtering the results using tag filters.

Fetch all metric definitions with tag filters

```
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:dc2,hostname:server1
```

```json
[
  {
    "id": "free_memory",
    "type": "gauge",
    "tags": {
      "datacenter": "dc2",
      "hostname": "server1"
    },
    "data": [
      {
        "timestamp": 1460711012361,
        "value": 2012
      }
    ]
  }
]
```

Fetch all metric definitions using RegExp tag filters

```
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:.*2
```

```json
[
  {
    "id": "free_memory",
    "type": "gauge",
    "tags": {
      "datacenter": "dc2",
      "hostname": "server1"
    },
    "data": [
      {
        "timestamp": 1460711012361,
        "value": 2012
      }
    ]
  }
]
```

#### Raw Data

The simplest form of querying for raw data points does not require any parameters and returns a list of data points.

```
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1460111000000"
```

Response with gauge data points

```json
[
  {
    "timestamp": 1460711012361,
    "value": 2012
  },
  {
    "timestamp": 1460671067863,
    "value": 2048
  },


  {
    "timestamp": 1460111065369,
    "value": 2048
  }
]

```

#### Date Range

Every query is bounded by a start and an end time. The end time defaults to now, and the start time defaults to 8 hours ago. These can be overridden with the start and end parameters respectively. The expected format of their values is a unix timestamp. The start of the range is inclusive while the end is exclusive.

```
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306"
```

#### Limiting Results

By default there is no limit on the number of data points returned. The limit parameter will limit the number of data points returned.

```
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&limit=3"
```

Response with 3 gauge data points

```json
[
  {
    "timestamp": 1460711012361,
    "value": 2012
  },
  {
    "timestamp": 1460671067863,
    "value": 2048
  },
  {
    "timestamp": 1460631064349,
    "value": 2012
  }
]
```

#### Aggregating Results using bucketDuration parameter

```
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&bucketDuration=600s"
```

```json
[
  {
    "start": 1460408700000,
    "end": 1460711100000,
    "empty": false,
    "samples": 8,
    "last": 2012,
    "avg": 2030,
    "sum": 16240
  },
  {
    "start": 1460106300000,
    "end": 1460408700000,
    "empty": false,
    "samples": 8,
    "last": 2012,
    "avg": 2030,
    "sum": 16240
  }
]
```

```
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&bucketDuration=600s&limit=1"
```

```json
[
  {
    "start": 1460408700000,
    "end": 1460711100000,
    "empty": false,
    "samples": 8,
    "last": 2012,
    "avg": 2030,
    "sum": 16240
  }
]
```

#### Query multi data time series

```
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d @query.json
```

query.json

```json
{
  "ids": ["free_memory"],
  "start": 1460111000000,
  "end": 1460711120000
}
```

result

```json
[
  {
    "id": "free_memory",
    "data": [
      {
        "timestamp": 1460111065369,
        "value": 2048
      },
      {
        "timestamp": 1460151065352,
        "value": 2012
      },


      {
        "timestamp": 1460711012361,
        "value": 2012
      }
    ]
  }
]
```
