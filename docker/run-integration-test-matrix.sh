#!/bin/bash

set -e
set -o pipefail

SCRIPT_DIR="$(dirname "$(readlink -f "$0")")"
ROOT_DIR="$(dirname "${SCRIPT_DIR}")"

kafkaVersions=`cat ${ROOT_DIR}/.github/workflows/lint_test.yml | yq -r '.jobs.integration_test.strategy.matrix.kafka_version | @csv'`

for kafkaVersionQuoted in $(echo $kafkaVersions | sed "s/,/ /g")
do
    kafkaVersion=`echo $kafkaVersionQuoted | sed 's/"//g'`
    cpVersion=`cat ${ROOT_DIR}/.github/workflows/lint_test.yml | sed -e "1,/kafka_version: ${kafkaVersion}/d" | head -n1 | xargs | cut -f2 -d ' '`

    echo "------------------------------------------------------------"
    echo " IT for Kafka $kafkaVersion (confluent platform: $cpVersion)"
    echo "------------------------------------------------------------"
    export CP_VERSION=$cpVersion
    export KAFKAVERSION=$kafkaVersion

    set +e
    $SCRIPT_DIR/run-integration-tests.sh
done
