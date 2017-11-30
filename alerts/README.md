# mohawk/alerts

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Alerting rules

Alerting rules in Mohawk servers send alerts to an Alertbuffer for processing.

## Configuring Alerts

Alerting Configuration is done using the config.yaml file.
Using the alerts key, we set a list of alerts, each alert has a unique name, and the alerts type.

For example:

```yaml
alerts:
- id: "free_memory is low"
  metric: "free_memory"
  # valid range for metric is set -
  # from > value <= to
  from: 1500
  to: 2000
  # type: 1 - alert if metric is out of valid range
  # type: 2 - alert if metric is below valid range
  # type: 3 - alert if metric is above valid range
  type: 1
- id: "free_memory is extremely low"
  metric: "free_memory"
  to: 1500
  type: 2
```
