#!/bin/sh

TARGET=$1
BIN_PATH=$2

if [ "linux_amd64" == "$TARGET" ]; then
  PATH=`echo ${BIN_PATH} |xargs dirname`
  echo "generating completions... $PATH"

  pushd $PATH
  echo "" > /tmp/empty.yaml
  ./kafkactl completion bash > ../../kafkactl-completion.bash --config-file=/tmp/empty.yaml
  popd
fi