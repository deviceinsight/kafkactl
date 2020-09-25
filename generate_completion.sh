#!/bin/sh

SCRIPT_PATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

TARGET=$1
BIN_PATH=$2

ARCH=linux_amd64

if [ "$(uname)" == "Darwin" ]; then
  ARCH=darwin_amd64
fi

if [ "$ARCH" == "$TARGET" ]; then
  echo "generating completions... ${BIN_PATH}"

  echo "" > /tmp/empty.yaml
  ${BIN_PATH} completion bash > ${SCRIPT_PATH}/kafkactl-completion.bash --config-file=/tmp/empty.yaml
fi
