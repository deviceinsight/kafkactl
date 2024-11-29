# Contributing to kafkactl

This guide should help you to get started with contribution to kafkactl.

It is still work in progress, but hopefully already addresses some typical questions.

## Testing your changes

_kafkactl_ is tested mainly with integration tests against kafka clusters in different
versions.

These tests typically run via Github Actions, but it is also easy to run them locally
on your development machine.

> :bulb: the integration tests do not clean up afterwards. To avoid interferences
> between different tests, we have helpers that create topics/groups/... with random suffixes. See e.g. [CreateTopic()](https://github.com/deviceinsight/kafkactl/blob/main/internal/testutil/helpers.go#L20)

### Run all tests locally

You can run all tests locally be executing:

```shell
./docker/run-integration-tests.sh
```

- This will start 3 kafka brokers, one zookeeper, an avro schema-registry.
- After the containers are up and ready, the tests are run.
- Finally, the containers are stopped.

### Run tests against different kafka cluster versions

In Github Actions we run the integration tests multiple times against different kafka cluster versions. In order to do the same locally just run:

```shell
./docker/run-integration-test-matrix.sh
```

### Develop tests locally

In order to develop integration tests locally, you just have to start the kafka cluster in docker:

```shell
cd docker
docker compose up -d
```

This will start 3 kafka brokers, one zookeeper, an avro schema-registry.

The tests will use [it-config.yml](https://github.com/deviceinsight/kafkactl/blob/main/it-config.yml) to connect to the cluster. You can check if you can connect to your cluster with:

```shell
kafkactl -V -C it-config.yml get brokers
```

This should give you some debug output followed by a list of three brokers.

That's all, now you can run integration tests directly in your IDE or from commandline.
