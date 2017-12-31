#!/bin/bash
mohawk -V &
while true
do
  # Post some fake data to mohawk
  timestamp=$(date +%s)
  value=$(($RANDOM % 100))
  curl -H "Content-Type: application/json" -X POST -d '[{"id": "free_memory", "data": [  {"timestamp":'"$timestamp"' , "value":'"$value"' } ]  }]' http://localhost:8080/hawkular/metrics/gauges/raw
  sleep 35
done
