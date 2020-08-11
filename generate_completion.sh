#!/bin/sh

TARGET=$1

if [ "linux_amd64" == "$TARGET" ]; then
  echo "generating completions..."

  echo "" > /tmp/empty.yaml
  ./kafkactl completion bash > dist/kafkactl-completion.bash --config-file=/tmp/empty.yaml
fi