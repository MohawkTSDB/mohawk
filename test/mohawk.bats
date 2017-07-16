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

@test "Mohawk is installed in version 0.19.4" {
  run mohawk -version

  [[ "$output" =~ "0.19.4"  ]]
}

@test "Server should be available" {
  wait_for_mohawk
  result="$(curl http://localhost:8080/hawkular/metrics/status)"
  kill_mohawk

  [[ "$result" =~ "STARTED" ]]
}
