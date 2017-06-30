#!/bin/bash

echo
echo "Metrics definition"
echo "------------------"
echo

echo "Get a list of metrics"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics | jq

echo "Post metrics data"
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d @data.json

echo "Put metrics tags"
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/tags -d @tags.json

echo "Get updated list of metrics"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics | jq

echo "Get a list of metrics taged datacenter:dc2,hostname:server1"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:dc2,hostname:server1 | jq

echo "Get a list of metrics taged using RegEx datacenter:.*2"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:.*2 | jq

echo
echo "Metrics query"
echo "-------------"
echo

echo "Query metrics"
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1460111000000&end=1460711120000" | jq

echo "Query metrics limit replay to 3 items"
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1460111000000&end=1460711120000&limit=3" | jq

echo "Query aggregated metrics"
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1460111000000&end=1460711120000&bucketDuration=302400s" | jq

echo "Query multi metrics timeseries"
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d @query.json | jq

echo
