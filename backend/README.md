

# mohawk/backend

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## Backends
Mohawk can use different backends for different use cases. Different backends may vary in speed, persistancy and scalability. Mohawk use a RESTful API identical to Hawkular, inheriting Hawkular's echosystem of clients and plugins.

## Backend Development

A backend should implement the [backend interface](/backend/backend.go). Each plugin is built for specific use case,
with features that best suite this use case. Implementation of a feature should not interfere
with plugin functionality, for example, a plugin built for speed may choose not to implement a feature that
may slow it down.

Plugins that implement a subset of the interface, must fail silently for unimplemented requests.

For more starting template for a plugin, look for the [backend example](/backend/example).

## Backend Features

|                  |               | Example | Memory        | Sqilte           |
|------------------|---------------|---------|---------------|------------------|
| Write to         |               |         | Local Memory  | Local File       |
| Speed            |               |         | Very Fast     | Fast             |
| Scale-ability    |               |         |               |                  |
| Retention        |               |         | 7 days        | File size        |
| Implements       | Multi Tenants |         | Y             | Y                |
|                  | Read          | Y       | Y             | Y                |
|                  | Write         |         | Y             | Y                |
|                  | Update        |         | Y             | Y                |
|                  | Delete        |         |               |                  |
