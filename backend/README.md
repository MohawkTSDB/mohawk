

# mohawk/backend

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## Backends

Mohawk can use different [backends](/backend) for different use cases. Different backends may vary in speed, persistancy and scalability. Mohawk use a subset of Hawkular's [RESTful API](/examples/REST.md), inheriting Hawkular's echosystem of clients and plugins.

## Backend Development

A backend should implement the [backend interface](/backend/backend.go). Each plugin is built for specific use case,
with features that best suite this use case. Implementation of a feature should not interfere
with plugin functionality, for example, a plugin built for speed may choose not to implement a feature that
may slow it down.

Plugins that implement a subset of the interface, must fail silently for unimplemented requests.

For a starting template of a plugin, look at the [backend example](/backend/example) directory.

## Backends Comparison

  - Example - a backend template.
  - Sqlite  - a file storage based backend.
  - Memory  - a memory storage based backend.

#### Features

|                  | Speed         | Retention Limit | Scaleability  | Storage          |
|------------------|---------------|-----------------|---------------|------------------|
| Example          |               |                 |               | No storage       |
| Memory           | Very Fast     | 7 days          |               | Memory           |
| Sqlite           | Fast          |                 |               | Local File       |

#### REST Endpoint Implementation

|                  | Multi Tenancy | Read| Write | Update | Delete |
|------------------|---------------|-----|-------|--------|--------|
| Example          |               | Y   |       |        |        |
| Memory           | Y             | Y   | Y     | Y      |        |
| Sqlite           | Y             | Y   | Y     | Y      | Y      |

#### Metrics List Implementation

|                  | Filter by Tag RegEx | Last Values |
|------------------|---------------------|-------------|
| Example          |                     |             |
| Memory           | Y                   | Y           |
| Sqlite           | Y                   |             |

#### Aggregation and Statistics Implementation

|                  | Min | Max| First | Last | Avg | Median | Std | Sum | Count |
|------------------|-----|----|-------|------|-----|--------|-----|-----|-------|
| Example          |     |    |       |      | Y   |        |     |     | Y     |
| Memory           |     |    |       | Y    | Y   |        |     | Y   | Y     |
| Sqlite           | Y   | Y  |       |      | Y   |        |     | Y   | Y     |
