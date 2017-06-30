#!/bin/bash

echo
echo "Metrics definition"
echo "------------------"
echo

echo
echo "Get a list of metrics"
echo "GET https://localhost:8443/hawkular/metrics/metrics"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics |jq '.'

echo
echo "Post metrics data"
echo "POST https://localhost:8443/hawkular/metrics/gauges/raw -d @data.json"
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw -d @data.json

echo
echo "Put metrics tags"
echo "PUT https://localhost:8443/hawkular/metrics/gauges/tags -d @tags.json"
curl -ks -X PUT https://localhost:8443/hawkular/metrics/gauges/tags -d @tags.json

echo
echo "Get updated list of metrics"
echo "GET https://localhost:8443/hawkular/metrics/metrics"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics |jq '.'

echo
echo "Get a list of metrics taged datacenter:dc2,hostname:server1"
echo "GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:dc2,hostname:server1"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:dc2,hostname:server1 |jq '.'

echo
echo "Get a list of metrics taged using RegEx datacenter:.*2"
echo "GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:.*2"
curl -ks -X GET https://localhost:8443/hawkular/metrics/metrics?tags=datacenter:.*2 |jq '.'

echo
echo "Metrics query"
echo "-------------"
echo

echo
echo "Query metrics"
echo "GET \"https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306\""
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306" |jq '.'

echo
echo "Query metrics limit replay to 3 items"
echo "GET \"https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&limit=3\""
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&limit=3" |jq '.'

echo
echo "Query aggregated metrics"
echo "GET \"https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&bucketDuration=600s\""
curl -ks -X GET "https://localhost:8443/hawkular/metrics/gauges/free_memory/raw?start=1498832548306&end=1498835518306&bucketDuration=600s" |jq '.'

echo
echo "Query multi metrics timeseries"
echo "POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d @query.json"
curl -ks -X POST https://localhost:8443/hawkular/metrics/gauges/raw/query -d @query.json |jq '.'

echo
