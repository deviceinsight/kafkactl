#!/bin/bash
# use:
# export NO_DOCKER_COMPOSE=true
# to skip docker compose when it is already running locally
set -e
set -o pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"

echo "using confluent platform version: ${CP_VERSION}"
echo "using kafka version: ${KAFKAVERSION}"

# docker compose up
pushd ${ROOT_DIR}
if [[ -z "${NO_DOCKER_COMPOSE}" ]]; then
  docker compose -f ${SCRIPT_DIR}/docker-compose.yml --env-file=${SCRIPT_DIR}/.env up -d
fi


# docker compose down
function tearDown {
  popd >/dev/null 2>&1
  if [[ -z "${NO_DOCKER_COMPOSE}" ]]; then
    docker compose -f ${SCRIPT_DIR}/docker-compose.yml --env-file=${SCRIPT_DIR}/.env down
  fi
}
trap tearDown EXIT

# wait for kafka to be ready
${SCRIPT_DIR}/wait-for-kafka.sh

# clear log file
[ -f integration-test.log ] && rm integration-test.log

# run integration tests
go tool gotestsum --format testname -- -run Integration ./...
