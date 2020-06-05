#!/bin/bash

set -e
set -o pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"

# docker compose up
pushd ${ROOT_DIR}
docker-compose -f ${SCRIPT_DIR}/docker-compose.yml up -d

# docker compose down
function tearDown {
  popd
  docker-compose -f ${SCRIPT_DIR}/docker-compose.yml down
}
trap tearDown EXIT

# wait for kafka to be ready
${SCRIPT_DIR}/wait-for-kafka.sh

# run integration tests
go test -v -run Integration ./...
