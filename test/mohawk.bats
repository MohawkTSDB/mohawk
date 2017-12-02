#!/usr/bin/env bats

NOW="$(($(date +%s) - 10000))000"

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
  curl http://localhost:8080/hawkular/metrics/gauges/raw -d "[{\"id\":\"free_memory\",\"data\":[{\"timestamp\":$NOW,\"value\": 42}]}]"
  result="$(curl http://localhost:8080/hawkular/metrics/gauges/free_memory/raw?limit=1)"
  kill_mohawk

  [ "$result" = "[{\"timestamp\":$NOW,\"value\":42}]" ]
}
