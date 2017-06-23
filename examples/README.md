

# mohawk/examples

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based [RESTful API](/examples/REST.md) as the primary interface.

## Usage

When installed, run using the command line ``mohawk``

```bash
$> mohawk --version
MoHawk version: 0.12.5
```

The `-h` flag will print out a help text, that list the command line arguments.

```
$> mohawk -h
Usage of ./mohawk:
  -backend string
    	the backend to use [sqlite, memory, example] (default "sqlite")
  -cert string
    	path to TLS cert file (default "server.pem")
  -gzip
    	accept gzip encoding
  -key string
    	path to TLS key file (default "server.key")
  -options string
    	specific backend options [e.g. db-dirname]
  -port int
    	server port (default 8080)
  -quiet
    	less debug output
  -tls
    	use TLS server
  -verbose
    	more debug output
  -version
    	version number
```

Running ``mohawk`` without ``tls`` and using the ``sqlite`` back end.

```bash
$> mohawk
2017/01/03 10:06:50 Start server, listen on http://0.0.0.0:8080
```

#### JSON + RESTful API
JSON over [RESTful API](/examples/REST.md) is the primary interface of MoHawk Metrics. This makes it easier for users to get started and also makes integration easier since REST+JSON is widely used and easily understood. a rich, growing set of features that includes:

#### Multi Tenancy
MoHawk Metrics provides virtual multi tenancy. All data is mapped to a tenant. Everything is partitioned by tenant. All requests, both reads and writes, can include a tenant id, default tenant id is "\_ops".

#### Tagging
MoHawk Metrics provides flexible tagging support that makes it easy to organize and group data. Tagging can also be used to provide additional information and context about data.

#### Querying
MoHawk Metrics offers a rich set of features around querying that are ideal for rendering data in graphs and in charts. This includes:

  - Filtering and grouping with tags
  - Searching metric definitions
  - Downsampling and aggregation
  - Limit and order results

#### Tenants

All data is partitioned by tenant. The partitioning happens at the API level. This means that a metric cannot exist on its own outside of a tenant.

#### Implicit tenant creation

```
curl -X POST http://localhost:8080/hawkular/metrics/gauges/raw -d @request.json -H "Hawkular-Tenant: com.acme"
```

This is a request to insert gauge data points for the com.acme tenant. If that tenant does not already exist, it will be request when storing the metric data. Specific details on inserting data can be found in Inserting Data.

#### Tenant Header

As previously stated all data is partitioned by tenant. MoHawk Metrics enforces this by allowing the Hawkular-Tenant HTTP header in requests. The value of the header is the tenant id. We saw this already with the implicit tenant creation.

Using the Hawkular-Tenant HTTP header in request:

```
curl http://localhost:8080/hawkular/metrics/metrics?tags=zone:us-west-1,kernel_version=4.0.9 -H "Hawkular-Tenant: com.acme"
```

#### Tenant Ids

A tenant has an id that uniquely identifies it. The id is a variable length, UTF-8 encoded string. MoHawk Metrics does not perform any validation checks to prevent duplicate ids. If the key already exists in the map, it will simply be overwritten with the new value.

#### Inserting Data

Inserting data is a synchronous operation with respect to the client. An HTTP response is not returned until all data points are inserted. On the server side however, multiple inserts to the database are done in parallel to achieve higher throughput.

#### Data Points

A data point in MoHawk Metrics is a tuple that in its simplest form consists of a timestamp and a value.

#### Timestamps

Timestamps are unix timestamps in milliseconds.

##### Insert data points

```
curl -X POST http://localhost:8080/hawkular/metrics/gauges/raw -d @request.json
```

request.json

```
[
  {
    "id": "free_memory",
    "data": [
      {"timestamp": 1460111065369, "value": 2048},
      {"timestamp": 1460151065369, "value": 2012}
    ]
  },
  {
    "id": "used_memory",
    "data": [
      {"timestamp": 1460111065369, "value": 2048},
      {"timestamp": 1460151065369, "value": 2075}
    ]
  }
]
```

Each array element is an object that has id and data properties. data contains an array of data points.

#### Tagging

Tags in MoHawk Metrics are key/value pairs. Tags can be applied to a metric to provide meta data for the time series as a whole. Tags can be used to perform filtering in queries.

#### Updating Metric Tags

These endpoints are used to add or replace tags.

```
curl -X PUT http://localhost:8080/hawkular/metrics/gauges/request_size/tags -d @tags.json
```

tags.json

```
{
  "datacenter": "dc2",
  "hostname": "server1"
}
```

#### Tag Filtering

MoHawk Metrics provides regular expression support for tag value filtering.

| Type           | Example       |                                                               |
|----------------|---------------|---------------------------------------------------------------|
| tag_name:regex | hostname:.*01 | Search for tag named hostname with a value that ends with 01. |

#### Querying

The examples provided in the following sections are not an exhaustive listing of the full API.

#### These operations do not fetch data points but rather the metric definition itself.

The next example illustrates the type parameter which filters the results by the specified types.

Fetch all metric definitions

```
curl http://localhost:8080/hawkular/metrics/metrics
```

response body

```
[
  {
    "id": "gauge_1"
    "type": "gauge"
  },
  {
    "id": "gauge_2",
    "type": "gauge"
  },
  {
    "id": "request_count",
    "type": "gauge"
  },
  {
    "id": "request_timeouts",
    "type": "gauge",
  }
]
```

The next example demonstrates querying metric and filtering the results using tag filters.

Fetch all metric definitions with tag filters

```
curl http://localhost:8080/hawkular/metrics/metrics?tags=zone:us-west-1,kernel_version=4.0.9
```

#### Raw Data

The simplest form of querying for raw data points does not require any parameters and returns a list of data points.

```
curl http://localhost:8080/hakwular/metrics/gauges/request_size/raw
```

Response with gauge data points

```
[
  {"timestamp": 1460413065369, "value": 3.14},
  {"timestamp": 1460212025569, "value": 4.57},
  {"timestamp": 1460111065369, "value": 5.056}
]
```

#### Date Range

Every query is bounded by a start and an end time. The end time defaults to now, and the start time defaults to 8 hours ago. These can be overridden with the start and end parameters respectively. The expected format of their values is a unix timestamp. The start of the range is inclusive while the end is exclusive.

```
curl http://localhost:8080/hawkular/metrics/gauges/request_size/raw?start=1235,end=6789
```

#### Limiting Results

By default there is no limit on the number of data points returned. The limit parameter will limit the number of data points returned.

```
curl http://localhost:8080/hawkular/metrics/gauges/request_size/raw?limit=100
```

#### Aggregating Results using bucketDuration parameter

```
curl http://localhost:8080/hawkular/metrics/gauges/request_size/raw?start=1235&end=6789&bucketDuration=60s
```

Response with gauge data points


```
[
  {
    "start": 6789,
    "end": 12345,
    "empty": false,
    "min": 0,
    "avg": 107.5,
    "max": 0,
    "median": 0,
    "sum": 0,
    "samples": 5
  },
  {
    "start": 12345,
    "end": 18345,
    "empty": false,
    "min": 0,
    "avg": 107.5,
    "max": 0,
    "median": 0,
    "sum": 0,
    "samples": 5
  },
]
```
