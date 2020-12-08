#!/bin/bash

set -e
set -o pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"

echo "using confluent platform version: ${CP_VERSION}"
echo "using kafka version: ${KAFKAVERSION}"

# docker compose up
pushd ${ROOT_DIR}
docker-compose -f ${SCRIPT_DIR}/docker-compose.yml --env-file=${SCRIPT_DIR}/.env up -d

# docker compose down
function tearDown {
  popd >/dev/null 2>&1
  docker-compose -f ${SCRIPT_DIR}/docker-compose.yml --env-file=${SCRIPT_DIR}/.env down
}
trap tearDown EXIT

# wait for kafka to be ready
${SCRIPT_DIR}/wait-for-kafka.sh

# clear log file
[ -f integration-test.log ] && rm integration-test.log

# run integration tests
go get gotest.tools/gotestsum
gotestsum --format testname -- -run Integration ./...
