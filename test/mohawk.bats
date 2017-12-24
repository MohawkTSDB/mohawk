#!/usr/bin/env bats

NOW="$(($(date +%s) - 10))000"

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
  cmd="curl http://localhost:8080/hawkular/alerts/raw?id=free_memory+is+low+or+high&state=T"

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

@test "Mohawk is installed in version 0.27.0" {
  run mohawk --version

  [[ "$output" =~ "0.27.0"  ]]
}

@test "Server should be available" {
  args=""

  wait_for_mohawk
  result="$(curl http://localhost:8080/hawkular/metrics/status)"
  kill_mohawk

  [[ "$result" =~ "STARTED" ]]
}

@test "Data post and get" {
  args=""

  wait_for_mohawk
  curl http://localhost:8080/hawkular/metrics/gauges/raw -d "[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW,\"value\": 42}]}]"
  result="$(curl http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)"
  kill_mohawk

  [ "$result" = "[{\"timestamp\":$NOW,\"value\":42}]" ]
}

@test "Reject requests when using bearer token" {
  args="--bearer-auth=123"

  wait_for_mohawk
  result1=$(curl -H "Authorization: Bearer 123" http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result2=$(curl -H "Authorization: Bearer 142" http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result3=$(curl http://localhost:8080/hawkular/metrics/status)
  kill_mohawk

  [ "$result1" = "[]" ]
  [[ "$result2" =~ "401" ]]
  [[ "$result3" =~ "STARTED" ]]
}

@test "Reject requests when using basic auth" {
  args="--basic-auth=user:pass"

  wait_for_mohawk
  result1=$(curl -H "Authorization: Basic dXNlcjpwYXNz" http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result2=$(curl -H "Authorization: Basic dXnLcjpwYXNz" http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result3=$(curl http://localhost:8080/hawkular/metrics/status)
  kill_mohawk

  [ "$result1" = "[]" ]
  [[ "$result2" =~ "401" ]]
  [[ "$result3" =~ "STARTED" ]]
}

@test "alerts are on" {
  args="--config=./src/alerts/examples/example.config.yaml"

  wait_for_mohawk
  result="$(curl http://localhost:8080/hawkular/alerts/status)"
  kill_mohawk

  [ "$result" = "{\"AlertsService\":\"STARTED\",\"AlertsInterval\":\"5s\",\"Heartbeat\":\"0\",\"ServerURL\":\"http://localhost:9099/append\"}" ]
}

@test "alerts fire" {
  args="--config=./src/alerts/examples/example.config.yaml --alerts-interval=1"
  data="[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW,\"value\":40}]}]"

  wait_for_mohawk
  result1="$(curl http://localhost:8080/hawkular/alerts/raw?id=free_memory+is+low+or+high)"
  curl http://localhost:8080/hawkular/metrics/gauges/raw -d "$data"
  wait_for_alert
  result2="$(curl http://localhost:8080/hawkular/alerts/raw?id=free_memory+is+low+or+high)"
  kill_mohawk

  [[ "$result1" =~ "\"State\":false" ]]
  [[ "$result2" =~ "\"State\":true" ]]
}
