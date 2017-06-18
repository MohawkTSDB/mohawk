# Introduction

MoHawk Metrics is a plug-able metrics storage engine that uses REST as the primary interface.

## REST
JSON over REST is the primary interface of MoHawk Metrics. This makes it easier for users to get started and also makes integration easier since REST+JSON is widely used and easily understood. a rich, growing set of features that includes:

### Multi Tenancy
MoHawk Metrics provides virtual multi tenancy. All data is mapped to a tenant. Everything is partitioned by tenant. All requests, both reads and writes, can include a tenant id, default tenant id is "\_ops".

### Tagging
MoHawk Metrics provides flexible tagging support that makes it easy to organize and group data. Tagging can also be used to provide additional information and context about data.

### Querying
MoHawk Metrics offers a rich set of features around querying that are ideal for rendering data in graphs and in charts. This includes:

    Filtering and grouping with tags

    Searching metric definitions

    Downsampling and aggregation

    Limit and order results

## Tenants

All data is partitioned by tenant. Data is not physically partitioned on disk. The partitioning happens at the API level. This means that a metric cannot exist on its own outside of a tenant. Let’s first look at how tenants are created.
Creating Tenants

Tenants are created in one of two ways. First, a tenant can be created implicitly by simply inserting metric data. Clients can immediately start storing data without first creating a tenant.

### Implicit tenant creation

```
curl -X POST http://server/hawkular/metrics/gauges/raw -d @request.json \
-H "Content-Type: application/json" -H "Hawkular-Tenant: com.acme"
```

This is a request to insert gauge data points for the com.acme tenant. If that tenant does not already exist, it will be request when storing the metric data. Specific details on inserting data can be found in Inserting Data.

### Tenant Header

As previously stated all data is partitioned by tenant. MoHawk Metrics enforces this by allowing the Hawkular-Tenant HTTP header in requests. The value of the header is the tenant id. We saw this already with the implicit tenant creation.

Using the Hawkular-Tenant HTTP header in request:

```
curl -X POST http://server/hawkular/metrics/counters?tags=zone:us-west-1,kernel_version=4.0.9 \
-H "Content-Type: application/json" -H "Hawkular-Tenant: com.acme"
```

### Tenant Ids

A tenant has an id that uniquely identifies it. The id is a variable length, UTF-8 encoded string. MoHawk Metrics does not perform any validation checks to prevent duplicate ids. If the key already exists in the map, it will simply be overwritten with the new value.

## Inserting Data

Inserting data is a synchronous operation with respect to the client. An HTTP response is not returned until all data points are inserted. On the server side however, multiple inserts to the database are done in parallel to achieve higher throughput.

### Data Points

A data point in MoHawk Metrics is a tuple that in its simplest form consists of a timestamp and a value. Timestamps are unix timestamps in milliseconds.

### Examples

There are several operations available for inserting data points.

#### Insert data points for multiple gauges

```
curl -X POST http://server/hawkular/metrics/gauges/raw -d @request.json \
-H "Content-Type: application/json"
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

## Tagging

Tags in MoHawk Metrics are key/value pairs. Tags can be applied to a metric to provide meta data for the time series as a whole. Tags can be used to perform filtering in queries.

### Updating Metric Tags

These endpoints are used to add or replace tags.

```
curl -X PUT http://server/hawkular/metrics/gauges/request_size/tags -d @tags.json \
-H "Content-Type: application/json"
```

tags.json

```
{
  "datacenter": "dc2",
  "hostname": "server1"
}
```

### Tag Filtering

MoHawk Metrics provides regular expression support for tag value filtering.

| Type           | Example       |                                                               |
|----------------|---------------|---------------------------------------------------------------|
| tag_name:regex | hostname:.*01 | Search for tag named hostname with a value that ends with 01. |

## Querying

The examples provided in the following sections are not an exhaustive listing of the full API.

### These operations do not fetch data points but rather the metric definition itself.

The next example illustrates the type parameter which filters the results by the specified types.

Fetch all metric definitions

```
curl -X POST http://server/hawkular/metrics/metrics \
-H "Content-Type: application/json"
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
    "type": "counter"
  },
  {
    "id": "request_timeouts",
    "type": "counter",
  }
]
```

The next example demonstrates querying metric and filtering the results using tag filters.

Fetch all metric definitions with tag filters

```
curl -X POST http://server/hawkular/metrics/metrics?tags=zone:us-west-1,kernel_version=4.0.9 \
-H "Content-Type: application/json"
```
