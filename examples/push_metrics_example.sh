#!/bin/bash

# Push random data into mohawk,
# requires mohawk server running on 127.0.0.1:8080
# e.g. mohawk -V &

usage="$(basename "$0") [-h] [-u -s -t -d -g] -- send random data into mohawk server

where:
    -h  show this help text
    -u  server url (default: http://127.0.0.1:8080)
    -t  tenant (default: _ops)
    -d  data id (default: free_memory)
    -g  data tags (default: node:localhost,group:usage)
    -s  sleep after each data push in sec (default: 30)

example:
    [ run mohawk: ./mohawk --options=granularity=1s --storage=memory ]

    # run push \"machine/example.com/cpu_usage\" metrics each 1s using the \"_system\" tenant
    # and add the labels \"name\" with value \"cpu_usage\" and \"hostname\" with value \"ecample.com\"
    ./push_metrics_example.sh -t _system -s 1 \\
        -d machine/example.com/cpu_usage \\
        -g name:cpu_usage,hostname:example.com
"

MOHAWK_URL=http://127.0.0.1:8080
MOHAWK_TENANT=_ops
DATA_ID=free_memory
SLEEP=30
TAGS=node:localhost,group:usage

while getopts u:s:d:g:t:h option
do
 case "${option}"
 in
 u) MOHAWK_URL=${OPTARG};;
 s) SLEEP=${OPTARG};;
 d) DATA_ID=${OPTARG};;
 g) TAGS=${OPTARG};;
 t) MOHAWK_TENANT=${OPTARG};;
 h) echo "$usage"
    exit;;
 \?) printf "illegal option: -%s\n" "$OPTARG" >&2
    echo "$usage" >&2
    exit 1;;
 esac
done

if [ -n "$TAGS" ]; then
  TAGS=\"$(echo $TAGS | sed 's/:/\":\"/g' | sed 's/,/\",\"/g')\"
  curl -k ${MOHAWK_URL}/hawkular/metrics/gauges/tags \
       -X PUT \
       -H "Content-Type: application/json" \
       -H "Hawkular-Tenant: ${MOHAWK_TENANT}" \
       -d "[{\"id\":\"${DATA_ID}\",\"tags\":{${TAGS}}}]"
fi

while true; do
  # Post some fake data to mohawk
  VAL=$(($RANDOM % 100))
  curl -k ${MOHAWK_URL}/hawkular/metrics/gauges/raw \
       -H "Content-Type: application/json" \
       -H "Hawkular-Tenant: ${MOHAWK_TENANT}" \
       -d "[{\"id\":\"${DATA_ID}\",\"data\":[{\"timestamp\":$(date +%s)000,\"value\":${VAL}}]}]"

  # Wait 30 sec
  sleep ${SLEEP}
done
