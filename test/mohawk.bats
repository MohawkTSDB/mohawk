#!/usr/bin/env bats

NOW="$(($(date +%s) - 10000))000"

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

@test "Mohawk binary is found in PATH" {
  run which mohawk

  [ "$status" -eq 0 ]
}

@test "Mohawk is installed in version 0.25.1" {
  run mohawk --version

  [[ "$output" =~ "0.25.1"  ]]
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
  args="--token=123"

  wait_for_mohawk
  result1=$(curl -H "Authorization: Bearer 123" http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result2=$(curl -H "Authorization: Bearer 142" http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)
  result3=$(curl http://localhost:8080/hawkular/metrics/status)
  kill_mohawk

  [ "$result1" = "[]" ]
  [[ "$result2" =~ "401" ]]
  [[ "$result3" =~ "STARTED" ]]
}

@test "Alerts are working" {
  mohawk --config="./alerts/examples/example.config.yaml" --verbose &
  # triggering alert number one.
  curl http://localhost:8080/hawkular/metrics/gauges/raw -d "[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW,\"value\": 12000}]}]"
  sleep 10 # allow the alerts worker to run..
  result="$(curl http://localhost:8080/hawkular/alerts/raw)"
  kill_mohawk
  [ "$result" = "[]" ]
}
