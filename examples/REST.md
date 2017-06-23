

# mohawk/examples/rest

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## RESTful API

#### Prefix: "/hawkular/metrics/"

| Method | Path           | Description             |
|--------|----------------|-------------------------|
| GET    | status         | Query server status     |
| GET    | tenants        | Query a list of tenants |
| GET    | metrics        | Query a list of metrics |

#### Prefix: "/hawkular/metrics/gauges/"

| Method | Path           | Description                    |
|--------|----------------|--------------------------------|
| GET    | :id/raw        | Query metric data              |
| POST   | raw            | Insert new metric data         |
| POST   | raw/query      | Query multiple metric data     |
| PUT    | :id/tags       | Update metric tags             |
