#!/bin/bash

usage="
Usage:
$(basename "$0") -f FROM -t TO [-i TENANT] [-j TENANT] [-g] [-h]  -- syndication script for Mohawk servers

where:
    -h  show this help text
    -f  from url e.g. https://metrics.node.com:8443
    -t  to url e.g. https://metrics.cluster.com:8443
    -i  from tenant e.g. _system [default to _ops]
    -j  to tenant e.g. _system   [default to use value of from tenant]
    -g  update tags [default is no]

example:
   $(basename "$0") -f https://localhost:8443 -t https://localhost:8444 -g -i _system"

while getopts hgf:t:i:j: option
do
 case "${option}"
 in
 h) echo "${usage}"; exit;;
 f) FROM_URL=${OPTARG};;
 t) TO_URL=${OPTARG};;
 i) FROM_TEN=${OPTARG};;
 j) TO_TENAN=${OPTARG};;
 g) TAGS="YES";;
 esac
done

if [ -z "${FROM_URL}" ]; then
  echo "${usage}"
  exit 1
fi

if [ -z "${TO_URL}" ]; then
  echo "${usage}"
  exit 1
fi

if [ -z "${FROM_TEN}" ]; then
  FROM_TEN="_ops"
fi

if [ -z "${TO_TENAN}" ]; then
  TO_TENAN="${FROM_TEN}"
fi

# set a tmp file for the source data
tmpfile=$(mktemp /tmp/mohawk-syndication.XXXXXX.json.gz)

# Get data from source
curl -ks -H "Accept-Encoding: gzip"  -H "Hawkular-Tenant: ${FROM_TEN}" -X GET  ${FROM_URL}/hawkular/metrics/metrics > "${tmpfile}"

# Post data to higher teir
curl -ks -H "Content-Encoding: gzip" -H "Hawkular-Tenant: ${TO_TENAN}" -X POST ${TO_URL}/hawkular/metrics/gauges/raw --data-binary "@${tmpfile}"

# if no tags, exit now
if [ -z "${TAGS}" ]; then
  rm "$tmpfile"
  exit
fi

# Post tags to higher teir
curl -ks -H "Content-Encoding: gzip" -H "Hawkular-Tenant: ${TO_TENAN}" -X PUT  ${TO_URL}/hawkular/metrics/gauges/tags --data-binary "@${tmpfile}"

rm "$tmpfile"
