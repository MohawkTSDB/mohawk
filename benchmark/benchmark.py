#!/bin/env python

from datetime import datetime, timedelta
from hawkular.metrics import HawkularMetricsClient, MetricType, datetime_to_time_millis
from random import uniform

tmilli = int(datetime_to_time_millis(datetime.utcnow()))
client = HawkularMetricsClient(tenant_id='python_test', auto_set_legacy_api=False)

# print server data
print client.query_status()

# write 1000 times
for i in xrange(1, 1000):
    t = tmilli - i * 1000 * 60
    v = uniform(0, 100)
    client.push(MetricType.Gauge, 'example.doc.1', v, t)

# read 1000 times
for i in xrange(1, 1000):
    t = tmilli - i * 1000 * 60
    v = client.query_metric(MetricType.Gauge, 'example.doc.1', start=t - 1000 * 90, end=t)
    print t - 1000 * 61, t, v
