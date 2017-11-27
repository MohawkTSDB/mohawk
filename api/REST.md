

# mohawk/usage/rest

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Examples

For usage information look at the [usage](/usage) directory.

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
	LastValues []DataItem        `json:"data,omitempty"`

#### DataItem

	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`

#### StatItem

	Start   int64   `json:"start"`
	End     int64   `json:"end"`
	Empty   bool    `json:"empty"`
	Samples int64   `json:"samples,omitempty"`
	Min     float64 `json:"min,omitempty"`
	Max     float64 `json:"max,omitempty"`
	First   float64 `json:"first,omitempty"`
	Last    float64 `json:"last,omitempty"`
	Avg     float64 `json:"avg,omitempty"`
	Median  float64 `json:"median,omitempty"`
	Std     float64 `json:"std,omitempty"`
	Sum     float64 `json:"sum,omitempty"`
