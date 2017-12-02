#!/usr/bin/env bats

wait_for_mohawk() {
  mohawk &

  for i in $(seq 0 50); do
    if [[ $(pidof mohawk) != "" ]]; then
      return 0
    fi
    sleep 0.1;
  done
}

kill_mohawk() {
  pkill -KILL mohawk || true
}

@test "Mohawk binary is found in PATH" {
  run which mohawk

  [ "$status" -eq 0 ]
}

@test "Mohawk is installed in version 0.22.1" {
  run mohawk --version

  [[ "$output" =~ "0.22.1"  ]]
}

@test "Server should be available" {
  wait_for_mohawk
  result="$(curl http://localhost:8080/hawkular/metrics/status)"
  kill_mohawk

  [[ "$result" =~ "STARTED" ]]
}

@test "Data post and get" {
  wait_for_mohawk
  curl -ks -X POST -H "Content-Type: application/json" -d '[{ "id": "free_memory", "data": [{"timestamp": 1498832548306, "value": 2075}]}]' http://localhost:8080/hawkular/metrics/gauges/raw
  result="$(curl -X GET http://localhost:8080/hawkular/metrics/metrics)"
  kill_mohawk
  [ "$result" = '[{"id":"free_memory","type":"gauge","tags":{},"data":[{"timestamp":1498832548306,"value":2075}]}]' ]
}