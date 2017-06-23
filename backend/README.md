

# mohawk/backend

![MoHawk](/images/logo-128.png?raw=true "MoHawk Logo")

MOck HAWKular, a Hawk[ular] with a mohawk, is a metrics storage engine that uses a plugin architecture for data storage and a Hawkular based RESTful API as the primary interface.

## Backends
Mohawk can use different backends for different use cases. Different backends may vary in speed, persistancy and scalability. Mohawk use a RESTful API identical to Hawkular, inheriting Hawkular's echosystem of clients and plugins.

  - Example backend, a template for new backends, generate random metrics.
  - Memory backend, speed: very fast, persistancy: while process is up (write to memory), scalability: no (write to memory)
  - Sqlite backend, speed: fast, persistancy: yes (write to file), scalability: no (write to file)
