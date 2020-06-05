#!/bin/bash

timeout=60
time=0

docker inspect wait-for-kafka 2>/dev/null | jq -c -e '.[0].State | select (.Status == "exited" and .ExitCode == 0)' > /dev/null

while [ $? -ne 0 ]; do
    echo "waiting for kafka..."
    sleep 5
    time=$(( time + 5 ))
    [[ time -ge $timeout ]] && echo "timeout waiting for kafka after ${timeout}s" && exit 1

    docker inspect wait-for-kafka 2>/dev/null | jq -c -e '.[0].State | select (.Status == "exited" and .ExitCode == 0)' > /dev/null
done

echo "kafka is ready."
