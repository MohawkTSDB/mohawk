# mohawk/alerts

![Mohawk](/images/logo-128.png?raw=true "Mohawk Logo")

Mohawk is a metric data storage engine that uses a plugin architecture for data storage and a simple REST API as the primary interface.

## Alerting rules

Alerting rules in Mohawk servers send alerts to an Alertbuffer if a metric value is outside valid range.

## Usage

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
{"ID":"free_memory is low","Metric":"free_memory","Tenant":"_ops","State":true,"AlertIfLowerThan":2000,"AlertIfHigherThen":8000,"Type":0,"TrigerValue":40,"TrigerTimestamp":1512142870000}
127.0.0.1 - - [01/Dec/2017 17:41:18] "POST /append HTTP/1.1" 200 -
{"ID":"free_memory is low","Metric":"free_memory","Tenant":"_ops","State":false,"AlertIfLowerThan":2000,"AlertIfHigherThen":8000,"Type":0,"TrigerValue":4000,"TrigerTimestamp":1512142893000}
127.0.0.1 - - [01/Dec/2017 17:41:38] "POST /append HTTP/1.1" 200 -
{"ID":"free_memory is low","Metric":"free_memory","Tenant":"_ops","State":true,"AlertIfLowerThan":2000,"AlertIfHigherThen":8000,"Type":0,"TrigerValue":40,"TrigerTimestamp":1512142901000}
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
- id: "free_memory is lower then 1000 or higher then 8000"
  metric: "free_memory"
  alert-if-lower-than: 1000
  alert-if-higher-than: 8000
- id: "free_memory is lower then 500"
  metric: "free_memory"
  alert-if-lower-than: 500
- id: "cpu_usage is above 95%"
  metric: "cpu_usage"
  alert-if-higher-than: 95
```
