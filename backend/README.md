

# mohawk/backend

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Storage Plugins

Mohawk can use different storage [plugins](/backend) for different use cases. Different storage backends may vary in speed, persistancy and scalability. Mohawk use a subset of Hawkular's [REST API](/examples/REST.md), inheriting Hawkular's echosystem of clients and plugins.

## Plugin Development

A storage plugin should implement the [backend interface](/backend/backend.go). Each storage plugin is built for specific use case, with features that best suite this use case.

Implementation of a feature should not interfere with the storage plugin functionality, for example, a plugin built for speed may choose not to implement a feature that may slow it down.

Plugins that implement a subset of the interface, must fail silently for unimplemented requests.

For a starting template of a storage plugin, look at the [backend example](/backend/example) directory.

## Plugins Comparison

  - Example - a backend template.
  - Sqlite  - a file storage based backend.
  - Memory  - a memory storage based backend.
  - Mongo   - a cluster based backend.

#### Features

| Plugin           | Speed         | Retention Limit | Scaleability  | Storage          |
|------------------|---------------|-----------------|---------------|------------------|
| Example          |               |                 |               | No storage       |
| Memory           | Very Fast     | 7 days          |               | Memory           |
| Sqlite           | Fast          |                 |               | Local File       |
| Mongo            | Fast          |                 | Cluster       | Mongo DB         |

#### REST Endpoint Implementation

| Plugin           | Multi Tenancy | Read| Write | Update | Delete |
|------------------|---------------|-----|-------|--------|--------|
| Example          |               | ✔️   |       |        |        |
| Memory           | ✔️             | ✔️   | ✔️     | ✔️      |        |
| Sqlite           | ✔️             | ✔️   | ✔️     | ✔️      | ✔️      |
| Mongo            | ✔️             | ✔️   | ✔️     | ✔️      |        |

#### Metrics List Implementation

| Plugin           | Filter by Tag RegEx | Last Values |
|------------------|---------------------|-------------|
| Example          |                     |             |
| Memory           | ✔️                   | ✔️           |
| Sqlite           | ✔️                   |             |
| Mongo            | ✔️                   |             |

#### Aggregation and Statistics Implementation

| Plugin           | Min | Max| First | Last | Avg | Median | Std | Sum | Count |
|------------------|-----|----|-------|------|-----|--------|-----|-----|-------|
| Example          |     |    |       |      | ✔️   |        |     |     | ✔️     |
| Memory           |     |    |       | ✔️    | ✔️   |        |     | ✔️   | ✔️     |
| Sqlite           | ✔️   | ✔️  |       |      | ✔️   |        |     | ✔️   | ✔️     |
| Mongo            | ✔️   | ✔️  | ✔️     | ✔️    | ✔️   |        |     | ✔️   | ✔️     |
