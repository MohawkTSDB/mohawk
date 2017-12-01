# mohawk/alerts

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Alerting rules

Alerting rules in Mohawk servers send alerts to an Alertbuffer if a metric value is outsde valid range.

## Runing Mohawk with alerting rules:

###### Running with some alerts rules in the config file:
```
./mohawk -c example.config.yaml
2017/12/01 17:40:58 Start server, listen on http://0.0.0.0:8080
...
```
###### Running the alert buffer:
```
./examples/alert-buffer.py
Starting httpd...
{"ID":"free_memory is low","Metric":"free_memory","Tenant":"_ops","State":true,"From":2000,"To":8000,"Type":0,"TrigerValue":40,"TrigerTimestamp":1512142870000}
127.0.0.1 - - [01/Dec/2017 17:41:18] "POST /append HTTP/1.1" 200 -
{"ID":"free_memory is low","Metric":"free_memory","Tenant":"_ops","State":false,"From":2000,"To":8000,"Type":0,"TrigerValue":4000,"TrigerTimestamp":1512142893000}
127.0.0.1 - - [01/Dec/2017 17:41:38] "POST /append HTTP/1.1" 200 -
{"ID":"free_memory is low","Metric":"free_memory","Tenant":"_ops","State":true,"From":2000,"To":8000,"Type":0,"TrigerValue":40,"TrigerTimestamp":1512142901000}
127.0.0.1 - - [01/Dec/2017 17:41:48] "POST /append HTTP/1.1" 200 -
...
```

###### Trigering the alerts using curl command:
```
$ curl http://localhost:8080/hawkular/metrics/gauges/raw -d "[{\"id\":\"free_memory\", \"data\":[{\"timestamp\": $(date +%s)000, \"value\": 4000}]}]"
$ curl http://localhost:8080/hawkular/metrics/gauges/raw -d "[{\"id\":\"free_memory\", \"data\":[{\"timestamp\": $(date +%s)000, \"value\": 40}]}]"
```

## Configuring Alerts

Alerting Configuration is done using the config.yaml file.
Using the alerts key, we set a list of alerts, each alert has a unique name, metric id and a valid range for that metric.
If a metric value change from being valid to not valid or from not valid to valid, an error status change is triggered and sent to allert buffer.

For example:

```yaml
alerts:
- id: "free_memory is low"
  metric: "free_memory"
  # valid range for metric is set -
  # from > value <= to
  # here valid range is from 1k to 8k, if free memory drops below 1k, error will be active.
  from: 1000
  to: 8000
  # type: 0 - alert if metric is out of valid range
  # type: 1 - alert if metric is above valid range
  # type: 2 - alert if metric is below valid range
  type: 0
- id: "free_memory is extremely low"
  metric: "free_memory"
  # here valid range is above 0.5k , if free memory drops below 0.5k, error will be active.
  from: 500
  # if type is 1 or 2, from or to values can be omitted.
  type: 2
- id: "cpu_usage is above 95%"
  metric: "cpu_usage"
  # here valid range is above 0% and below 95%, if cpu usage is above 95%, error will be active.
  from: 0
  to: 95
  # Default type is 0, it can be omitted.
```
