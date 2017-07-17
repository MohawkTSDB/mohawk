

# mohawk/examples/rest

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Examples

For usage information look at the [example](/examples) directory.

## REST API

#### Prefix: "/hawkular/metrics/"

| Method | Path           | Description             | Response Type    |
|--------|----------------|-------------------------|------------------|
| GET    | status         | Query server status     | Object           |
| GET    | tenants        | Query a list of tenants | Array of Strings |
| GET    | metrics        | Query a list of metrics | Array of Items   |

#### Prefix: "/hawkular/metrics/gauges/"

| Method | Path           | Description                    | Response Type                   |
|--------|----------------|--------------------------------|---------------------------------|
| GET    | :id/raw        | Query metric data              | Array of DataItems or StatItems |
| POST   | raw/query      | Query multiple metric data     | Array of DataItems or StatItems |
| PUT    | :id/tags       | Update metric tags             |                                 |
| PUT    | tags           | Update multiple metric tags    |                                 |
| POST   | raw            | Insert new metric data         |                                 |

## Data Structures

#### Item

	Id         string            `json:"id"`
	Type       string            `json:"type"`
	Tags       map[string]string `json:"tags"`
	LastValues []DataItem        `json:"data"`

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
	First   float64 `json:"first"`
	Last    float64 `json:"last"`
	Avg     float64 `json:"avg"`
	Median  float64 `json:"median"`
	Std     float64 `json:"std"`
	Sum     float64 `json:"sum"`
