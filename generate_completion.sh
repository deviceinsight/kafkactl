#!/bin/sh

SCRIPT=$(readlink -f "$0")
SCRIPT_PATH=$(dirname "$SCRIPT")

TARGET=$1
BIN_PATH=$2

if [ "linux_amd64" == "$TARGET" ]; then
  echo "generating completions... ${BIN_PATH}"

  echo "" > /tmp/empty.yaml
  ${BIN_PATH} completion bash > ${SCRIPT_PATH}/kafkactl-completion.bash --config-file=/tmp/empty.yaml
fi