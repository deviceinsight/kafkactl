#!/bin/sh

SCRIPT_PATH=`readlink -m $0/..`

TARGET=$1
BIN_PATH=$2

ARCH=linux_amd64

if [ "${BIN_PATH}" == "" ]; then
  BIN_PATH=${SCRIPT_PATH}/kafkactl
fi

if [ "$TARGET" == "" ]; then
  TARGET=$ARCH
fi

if [ "$(uname)" == "Darwin" ]; then
  ARCH=darwin_amd64
fi

if [ "$ARCH" == "$TARGET" ]; then
  echo "generating completions... ${BIN_PATH}"

  echo "" > /tmp/empty.yaml
  ${BIN_PATH} completion bash > ${SCRIPT_PATH}/kafkactl-completion.bash --config-file=/tmp/empty.yaml
else
  echo "not generating: $ARCH $TARGET"
fi
