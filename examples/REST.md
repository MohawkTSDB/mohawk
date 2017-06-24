

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

## Data Structures

#### Item

	Id   string            `json:"id"`
	Type string            `json:"type"`
	Tags map[string]string `json:"tags"`

#### DataItem

	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`

#### StatItem

	Start   int64   `json:"start"`
	End     int64   `json:"end"`
	Empty   bool    `json:"empty"`
	Samples int64   `json:"samples"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
	Median  float64 `json:"median"`
	Sum     float64 `json:"sum"`
