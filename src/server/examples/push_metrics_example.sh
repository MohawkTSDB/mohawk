#!/bin/bash

# Push random data into mohawk,
# requires mohawk server running on 127.0.0.1:8080
# e.g. mohawk -V &

MOHAWK_URL=127.0.0.1:8080
SLEEP=30

while true; do
  # Post some fake data to mohawk
  VAL=$(($RANDOM % 100))
  curl http://${MOHAWK_URL}/hawkular/metrics/gauges/raw \
       -H "Content-Type: application/json" \
       -d "[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":\"$(date +%s%N | cut -b1-13)\",\"value\":\"${VAL}\"}]}]"

  # Wait 30 sec
  sleep ${SLEEP}
done
