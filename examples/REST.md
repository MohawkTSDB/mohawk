

# mohawk/examples

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## RESTful API

### Prefix: "/hawkular/metrics/"

| Method | Path     | Description            | Arguments    | Example                  |
|--------|----------|------------------------|--------------|--------------------------|
| GET    | status   | Read server status     |              |                          |
| GET    | tenants  | Read a list of tenants |              |                          |
| GET    | metrics  | Read a list of metrics |              |                          |

### Prefix: "/hawkular/metrics/gauges/"

| Method | Path           | Description           | Arguments    | Example                  |
|--------|----------|------------------------|--------------|--------------------------|
| GET    | :id/raw        | Read metric data      |              |                          |
| POST   | raw            | Write new metric data |              |                          |
| POST   | raw/query      | Query metric data     |              |                          |
| PUT    | :id/tags       | Update metric tags    |              |                          |
| DELETE | :id/raw        |                       |              |                          |
| DELETE | :id/tags/:tags |                       |              |                          |
