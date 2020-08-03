#!/bin/sh

TARGET=$1
PATH=$2

if [ "linux_amd64" == "$TARGET" ]; then
  echo "generating completions... ($PATH)"
  echo "" > /tmp/empty.yaml
  $PATH completion bash > kafkactl-completion.bash --config-file=/tmp/empty.yaml
fi