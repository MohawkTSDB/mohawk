#!/usr/bin/env bats

NOW="$(($(date +%s) - 10))000"
NOW_5MN="$(($(date +%s) - 5 * 60))000"

wait_for_mohawk() {
  mohawk $args &

  for i in $(seq 0 50); do
    if [[ $(pidof mohawk) != "" ]]; then
      return 0
    fi
    sleep 0.1;
  done
}

kill_mohawk() {
  killall mohawk || true
}

wait_for_alert() {
  cmd="curl http://127.0.0.1:8080/hawkular/alerts/raw?id=free_memory+is+low+or+high&state=T"

  for i in $(seq 0 50); do
    if [[ $($cmd) != "[]" ]]; then
      return 0
    fi
    sleep 0.1;
  done
}

@test "Mohawk binary is found in PATH" {
  run which mohawk

  [ "$status" -eq 0 ]
}

@test "Mohawk is installed in version 0.28.2" {
  run mohawk --version

  [[ "$output" =~ "0.28.2"  ]]
}

@test "Server should be available" {
  args=""

  wait_for_mohawk
  result="$(curl http://127.0.0.1:8080/hawkular/metrics/status)"
  kill_mohawk

  [[ "$result" =~ "STARTED" ]]
}

@test "Data post and get" {
  args=""

  wait_for_mohawk
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw -d "[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW,\"value\": 42}]}]"
  result="$(curl http://127.0.0.1:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)"
  kill_mohawk

  [ "$result" = "[{\"timestamp\":$NOW,\"value\":42}]" ]
}

@test "Reject requests when using bearer token" {
  args="--bearer-auth=123"

  wait_for_mohawk
  result1=$(curl -H "Authorization: Bearer 123" http://127.0.0.1:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result2=$(curl -H "Authorization: Bearer 142" http://127.0.0.1:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result3=$(curl http://127.0.0.1:8080/hawkular/metrics/status)
  kill_mohawk

  [ "$result1" = "[]" ]
  [[ "$result2" =~ "401" ]]
  [[ "$result3" =~ "STARTED" ]]
}

@test "Reject requests when using basic auth" {
  args="--basic-auth=user:pass"

  wait_for_mohawk
  result1=$(curl -H "Authorization: Basic dXNlcjpwYXNz" http://127.0.0.1:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result2=$(curl -H "Authorization: Basic dXnLcjpwYXNz" http://127.0.0.1:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result3=$(curl http://127.0.0.1:8080/hawkular/metrics/status)
  kill_mohawk

  [ "$result1" = "[]" ]
  [[ "$result2" =~ "401" ]]
  [[ "$result3" =~ "STARTED" ]]
}

@test "alerts are on" {
  args="--config=./src/alerts/examples/example.config.yaml"

  wait_for_mohawk
  result="$(curl http://127.0.0.1:8080/hawkular/alerts/status)"
  kill_mohawk

  [ "$result" = "{\"AlertsService\":\"STARTED\",\"AlertsInterval\":\"5s\",\"Heartbeat\":\"0\",\"ServerURL\":\"http://localhost:9099/append\"}" ]
}

@test "alerts fire" {
  args="--config=./src/alerts/examples/example.config.yaml --alerts-interval=1"
  data="[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW,\"value\":40}]}]"

  wait_for_mohawk
  result1="$(curl http://127.0.0.1:8080/hawkular/alerts/raw?id=free_memory+is+low+or+high)"
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw -d "$data"
  wait_for_alert
  result2="$(curl http://127.0.0.1:8080/hawkular/alerts/raw?id=free_memory+is+low+or+high)"
  kill_mohawk

  [[ "$result1" =~ "\"State\":false" ]]
  [[ "$result2" =~ "\"State\":true" ]]
}

@test "parse relative time" {
  data="[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW_5MN,\"value\":40}]}]"
  query1="{\"ids\":[\"free_memory\"],\"start\":\"-2mn\"}"
  query2="{\"ids\":[\"free_memory\"],\"start\":\"-8mn\"}"

  wait_for_mohawk
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw -d "$data"
  result1="$(curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw/query -d $query1)"
  result2="$(curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw/query -d $query2)"
  kill_mohawk

  [[ "$result1" =~ "[]" ]]
  [[ "$result2" =~ "\"value\":40" ]]
}

@test "query metrics by tags" {
  data="[{\"id\":\"rss\",\"data\":[{\"timestamp\":$NOW_5MN,\"value\":40}],\"tags\":{\"name\":\"free_memory\"}}]"
  query="{\"tags\":\"name:free_memory\",\"start\":\"-8mn\"}"

  wait_for_mohawk
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw -d "$data"
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/tags -d "$data" -X PUT
  result="$(curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw/query -d $query)"
  kill_mohawk

  [[ "$result" =~ "\"value\":40" ]]
}

@test "query the m metrics endpoint by tags" {
  data="[{\"id\":\"rss\",\"data\":[{\"timestamp\":$NOW_5MN,\"value\":40}],\"tags\":{\"name\":\"free_memory\"}}]"
  query="{\"tags\":\"name:free_memory\",\"start\":\"-8mn\"}"

  wait_for_mohawk
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/raw -d "$data"
  curl http://127.0.0.1:8080/hawkular/metrics/gauges/tags -d "$data" -X PUT
  result="$(curl http://127.0.0.1:8080/hawkular/metrics/m/stats/query -d $query)"
  kill_mohawk

  [[ "$result" =~ "\"rss\":" ]]
}
