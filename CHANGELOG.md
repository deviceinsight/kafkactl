# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed
- Update to Go 1.24.12

## 5.17.0 - 2026-01-07
- [#309](https://github.com/deviceinsight/kafkactl/issues/309) Allow executing a script to retrieve a OAuth token.

- new flags `--filter-key`, `--filter-value` and `--filter-header` for filtering key/value/headers of messages

## 5.16.0 - 2025-12-03
- [#306](https://github.com/deviceinsight/kafkactl/pull/306) Support building arm image by updating to goreleaser v2
- new flag `--print-all` for consume command which combines the other print parameters into one
- decoding of AMQP-encoded header values

## 5.15.0 - 2025-11-13

## 5.14.0 - 2025-10-21

### Added
- [#304](https://github.com/deviceinsight/kafkactl/pull/304) Add support for headers when using file input to produce
- [#300](https://github.com/deviceinsight/kafkactl/issues/300) Allow mounting secrets with certificates to kafkactl pod
- [#253](https://github.com/deviceinsight/kafkactl/issues/253) Allow configuring k8s resource requests/limits

## 5.13.0 - 2025-09-22

### Added
- [#297](https://github.com/deviceinsight/kafkactl/pull/297) Support configuring SASL version

## 5.12.1 - 2025-08-27

### Fixed
- [#295](https://github.com/deviceinsight/kafkactl/issues/295) make sure we match the FullName first before using shortname

## 5.12.0 - 2025-08-26

### Added
- **Alter Broker Command**: New `kafkactl alter broker` command for dynamic broker configuration management without requiring broker restarts
- Only warn when config file or context file cannot be created (#282)
- **SCRAM User Management**: Complete SCRAM credential management with `kafkactl create user`, `kafkactl alter user`, `kafkactl delete user`, `kafkactl get users`, and `kafkactl describe user` commands supporting SCRAM-SHA-256 and SCRAM-SHA-512 mechanisms
- `describe broker` now allows to print also default broker configs

## 5.11.1 - 2025-07-28

## 5.11.0 - 2025-07-16

### Added
- [#276](https://github.com/deviceinsight/kafkactl/issues/276) More granular control when listing/deleting acls

### Fixed
- Do not commit offset for messages over the max-messages

## 5.10.1 - 2025-07-02

### Fixed
- gracefully shutdown Kubernetes pod on exit
- make plugin search path configurable via `KAFKA_CTL_PLUGIN_PATHS` environment variable

## 5.10.0 - 2025-06-24

### Added
- Current context is now stored in a separate file to avoid write operations on the config file.
- run kubectl commands as user other.

### Changed
- Update to Go 1.24.4

### Fixed
- Fixed panic() when configuring a nil map as tokenProvider.options

## 5.9.0 - 2025-06-16
### Fixed
- [#255](https://github.com/deviceinsight/kafkactl/pull/255) Set resource name when creating a cluster ACL

### Fixed
- SASL token provider plugin options are now working in kubernetes mode.

## 5.8.0 - 2025-05-16

## 5.7.0 - 2025-04-02

### Added
- [#232](https://github.com/deviceinsight/kafkactl/pull/232) Protobufs serialization and deserialization now supports schemaRegistry

## Changed
- Update to Go 1.24

## 5.6.0 - 2025-03-11

### Changed
- [#240](https://github.com/deviceinsight/kafkactl/pull/240) deserialization has been changed to generalize the usage
  of the schemaRegistry. Note that the configuration of the schemaRegistry changed with this.

## 5.5.1 - 2025-03-05

### Fixed
- [#237](https://github.com/deviceinsight/kafkactl/issues/237) Fix handling of relative paths.

## 5.5.0 - 2025-02-11
### Added
- [#234](https://github.com/deviceinsight/kafkactl/pull/234) caching to arvo client when producing messages
- [#236](https://github.com/deviceinsight/kafkactl/issues/236) set working directory to path with first loaded config file.
- [#233](https://github.com/deviceinsight/kafkactl/issues/233) replication factor is now printed for most topic output formats.

### Removed
- [#231](https://github.com/deviceinsight/kafkactl/issues/231) Remove support for installing kafkactl via snap.

### Fixed
- [#227](https://github.com/deviceinsight/kafkactl/issues/227) Incorrect handling of Base64-encoded values when producing from JSON
- [#228](https://github.com/deviceinsight/kafkactl/issues/228) Fix parsing version of gcloud kubectl.
- [#217](https://github.com/deviceinsight/kafkactl/issues/217) Allow reset of consumer-group in dead state

## 5.4.0 - 2024-11-28
### Added
- [#215](https://github.com/deviceinsight/kafkactl/pull/215) Add `--context` option to set command's context

## 5.3.0 - 2024-08-14
### Added
- [#203](https://github.com/deviceinsight/kafkactl/pull/203) Add pod override fields affinity and tolerations
- [#210](https://github.com/deviceinsight/kafkactl/pull/210) Create topic from file

## 5.2.0 - 2024-08-08

## 5.1.0 - 2024-08-07
- [#209](https://github.com/deviceinsight/kafkactl/pull/209) Allow configuring basicAuth for avro schema registry

### Added
- [#207](https://github.com/deviceinsight/kafkactl/pull/207) Allow configuring TLS for avro schema registry
- [#193](https://github.com/deviceinsight/kafkactl/pull/193) Print group instance IDs in `describe consumer-group` command

## 5.0.6 - 2024-03-14

### Added
- [#190](https://github.com/deviceinsight/kafkactl/pull/190) Improve handling of project config files
- [#192](https://github.com/deviceinsight/kafkactl/pull/192) Plugin infrastructure for tokenProviders

### Fixed
- [#151](https://github.com/deviceinsight/kafkactl/issues/151) Get topics command should not require describe permission

## 4.0.0 - 2024-01-18

### Added
- [#182](https://github.com/deviceinsight/kafkactl/issues/182) Make isolationLevel configurable and change default to 'readCommitted'
- [#184](https://github.com/deviceinsight/kafkactl/pull/184) Added option to show default configs when describing topics
- [#183](https://github.com/deviceinsight/kafkactl/issues/183) Add command `delete records` to delete records from topic

### Changed
- [#180](https://github.com/deviceinsight/kafkactl/issues/180) Change default for replication-factor when creating topics

## 3.5.1 - 2023-11-10

## 3.5.0 - 2023-11-10

### Added
- [#171](https://github.com/deviceinsight/kafkactl/issues/171) Support for reset command to reset offset to offset from provided datetime
- [#172](https://github.com/deviceinsight/kafkactl/issues/172) Support for json format in produce command

## 3.4.0 - 2023-09-25

### Fixed
- [#165](https://github.com/deviceinsight/kafkactl/issues/165) Kubernetes pod name now always includes random suffix
- [#165](https://github.com/deviceinsight/kafkactl/issues/158) fix parsing array params in kubernetes mode

## 3.3.0 - 2023-09-05
### Fixed
- [#162](https://github.com/deviceinsight/kafkactl/pull/162) Fix `consumer-group` crashes when a group member has no assignment
- [#160](https://github.com/deviceinsight/kafkactl/pull/160) Fix kubectl version detection

### Added
- [#159](https://github.com/deviceinsight/kafkactl/issues/159) Add ability to read config file from `$PWD/kafkactl.yml`

## 3.2.0 - 2023-08-17

## 3.1.0 - 2023-03-06
### Added
- [#121](https://github.com/deviceinsight/kafkactl/issues/121) Support message consumption with a specified timestamp range


## 3.0.3 - 2023-02-01
### Changed
- [#144](https://github.com/deviceinsight/kafkactl/issues/144) Dependencies have been updated

## 3.0.2 - 2023-01-20
### Changed
- [#143](https://github.com/deviceinsight/kafkactl/pull/143) Dependencies have been updated
- [#139](https://github.com/deviceinsight/kafkactl/pull/139) Option to print partitions in default output format

## 3.0.1 - 2022-11-08
### Fixed
- [#138](https://github.com/deviceinsight/kafkactl/issues/138) remove default value for `BROKERS` in Dockerfiles

## 3.0.0 - 2022-09-30
### Changed
- [#123](https://github.com/deviceinsight/kafkactl/issues/123) Make avro json codec configurable and switch default to standard json

## 2.5.0 - 2022-08-19
### Added
- [#105](https://github.com/deviceinsight/kafkactl/issues/105) Add replication factor to `get topics`
- Add `clone` command for `consumer-group` and `topic`

### Fixed
- [#108](https://github.com/deviceinsight/kafkactl/issues/108) Fix "system root pool not available on Windows" by migrating to go 1.19
- Print topic configs when yaml and json output format used with `-c` flag

## 2.4.0 - 2022-06-23
### Fixed
- [#129](https://github.com/deviceinsight/kafkactl/issues/129) Fix tombstone record handling if protobuf encoding used.
- allow producing messages where only key is protobuf encoded

### Changed
- Implicitly extend `--proto-import-path` with existing file path from `--proto-file` so that the parameter can be neglected in some cases.

## 2.3.0 - 2022-05-31

### Added
- [#127](https://github.com/deviceinsight/kafkactl/issues/127) the commands `create consumer-group` and `reset consumer-group-offset` can now be called with multiple `--topic` parameters.
  Additionally, `reset consumer-group-offset` can be called with `--all-topics` parameter, which will reset all topics in the group.

## 2.2.1 - 2022-04-04

### Fixed
- [#124](https://github.com/deviceinsight/kafkactl/issues/124) make sure input buffer can hold long messages when producing messages to a topic

## 2.2.0 - 2022-03-04

### Added
- add parameter `--max-messages` to consume command, to stop consumption after a fixed amount of messages
- [#60](https://github.com/deviceinsight/kafkactl/issues/60) add parameter `--group` to consume command, which allows consuming as a member of the given consumer group

## 2.1.0 - 2022-01-28

### Added
- add config options `image`, `imagePullSecret` in order to pull kafkactl from private registry when running in k8s (fixes #116).

## 2.0.1 - 2022-01-24

### Fixed
- make sure kafkactl pod name is always lowercase only

## 2.0.0 - 2022-01-17

### Added
- [#112](https://github.com/deviceinsight/kafkactl/issues/112) make `maxMessageBytes` configurable in produce command.

  :warning: this is a breaking change since the format of the `config.yml` has been changed in order to group producer
  related configurations under `context.producer`.

### Fixed
- fixed error handling for describe topic command. previously errors for requests to describe partitions had been swallowed.

### Changed
- changed naming of kafkactl pod when running in k8s. pod is now named based on clientID specified in config.yml

## 1.24.0 - 2021-12-03

### Added
- [#110](https://github.com/deviceinsight/kafkactl/issues/110) Added protobuf support to consume/produce commands (@xakep666 thank you for the contribution)

## 1.23.1 - 2021-11-23

### Fixed
 - fixed parsing of `--replicas` flag for `alter partition` when running against kubernetes cluster

## 1.23.0 - 2021-11-05

### Added
- Add new command `describe broker` to view broker configurations

### Fixed
- [#106](https://github.com/deviceinsight/kafkactl/issues/106) Do not check if topic is available before produce, because otherwise autoTopicCreation does not work

## 1.22.1 - 2021-10-27

- fix brew warning (https://github.com/goreleaser/goreleaser/pull/2591)

## 1.22.0 - 2021-10-14

### Added
- Added parameter `--null-value` to produce command in order to produce a null value (tombstone)

## 1.21.0 - 2021-10-01

### Added
- Added new command `delete consumer-group-offset` to delete a consumer-group-offset
- Add new command `get brokers` to get the list of brokers advertised by Kafka

## 1.20.1 - 2021-09-24

### Fixed
- SASL mechanism setting is now supported in `kafkactl attach`

## 1.20.0 - 2021-08-06

### Fixed
- Filter -C and --config-file option in kubernetes context

### Added
- Support `delete` command when accessing a remote cluster deployed on kubernetes

## 1.19.0 - 2021-07-26

### Added
- Added new commands `delete consumer-group` to delete a consumer-group

### Fixed
- Calls to the AVRO schema registry are now cached correctly

## 1.18.1 - 2021-07-15

### Fixed
- alter topic should fail, when trying to set number of partitions to the current number of partitions

## 1.18.0 - 2021-07-14

### Changed
- the default for required-acks in the producer has been changed to WaitForLocal in order to align with the JVM Producer.
  In addition, this setting is now also configurable.

## 1.17.2 - 2021-06-09

### Fixed
- fix version prefix in aur package. kubernetes support not working without the prefix.

## 1.17.1 - 2021-05-17

### Fixed
- Compatibility with new versions of `kubectl` by removing deprecated parameter from the command
- Replace `CMD` with `ENTRYPOINT` in Ubuntu Dockerfile to restore behavior documented in readme

## 1.17.0 - 2021-04-06

### Fixed
- SASL mechanism support now also implemented for cluster admin
- process no longer gets stuck when deserialization error occurs

### Added
- calls to avro schema registry are now cached in memory to improve performance

### Changed
- Updated avro library to latest version (v2.10.0)

## 1.16.0 - 2021-02-25

### Added
- Introduce configuration options to configure `sasl.mechanism` in order to use `scram-sha512` or `scram-sha256` instead
of `plaintext`.
- Added new commands `get acl`, `create acl`, `delete acl` to manage Kafka ACLs with kafkactl

## 1.15.1 - 2021-01-20

### Fixed
- Fixed Ctrl+C not working to stop kafkactl

## 1.15.0 - 2021-01-19

### Added
- Add parameter `--skip-empty` to `describe topic` command which allows skipping/hiding empty partitions.

### Changed
- `describe topic` now includes `EMPTY` field, which directly shows if a partition contains messages.

## 1.14.0 - 2020-12-08

### Added
- Add parameter `--replication-factor` to `alter topic` command which allows changing the replication factor of a topic.
  Note that kafka >= 2.4.0.0 is required, otherwise the relevant api calls are not available.
- Added command `alter partition` which currently only enables to manually assign broker replicas to a partition.
  Note that kafka >= 2.4.0.0 is required, otherwise the relevant api calls are not available.
- Added `requestTimeout` config to control timeout of admin requests.

## 1.13.3 - 2020-11-11

### Fixed
- Add certificates to TlsConfig when tls.insecure = true

## 1.13.2 - 2020-11-05

### Fixed
- `get consumer-groups -o compact --topic=xyz` no longer ignores the topic parameter

## 1.13.1 - 2020-10-30

## 1.13.0 - 2020-10-21

### Added
- Ubuntu Docker image has CA certificates installed.
- Scratch Docker image has CA certificates installed.

### Changed
- `get consumer-groups` now includes the assigned topics as long as output format is not `compact`.

## 1.12.0 - 2020-09-25

### Changed
- TLS configuration now starts off with the System's CA pool instead of a completely empty one. This improves support for AWS MSK with PCAs.

## 1.11.1 - 2020-08-11

## 1.11.0 - 2020-08-07

### Added
- direct support for kafka clusters [running in kubernetes](https://github.com/deviceinsight/kafkactl/blob/main/README.adoc#running-in-kubernetes)
- `attach` command to get a bash into kafkactl pod when running in kubernetes

## 1.10.0 - 2020-08-03

### Fixed
- auto-completion should now work consistent for all supported shells and provides dynamic completion for
 e.g. names of topics or consumer-groups.

### Added
- Add parameters `--key-encoding`, `--value-encoding` to produce command to write messages from hex/base64
- Add parameters `--key-encoding`, `--value-encoding` to consume command to print messages as hex/base64
- auto-completion now also works inside ubuntu docker image
- bash auto-completion should work out of the box, when kafkactl is installed via snap

### Changed
- improved and documented overriding config keys via environment variables
- Binary (non-string) messages are auto-detected and printed as base64 by default

## 1.9.0 - 2020-06-19

### Added
- Add parameter `--header` to produce command to include message headers when writing messages

### Changed
- generation of commands and error handling have been refactored in order to allow for better testability
- dockerfile is build from UBUNTU:LATEST instead of SCRATCH

### Fixed
- loading custom CA cert failed in 1.8.0

## 1.8.0 - 2020-05-13

### Added
- add configuration option `sasl` for connection with SASL PLAIN username and password
- show oldestOffset,lead values when describing consumer groups
- allow omitting the CA cert
- show error message when trying to reset offset of a non-empty group
- allow enabling tls without specifying additional options

## 1.7.0 - 2020-03-05

### Added
- Add configuration option `tlsInsecure` to be able to connect with TLS to IP address served brokers.
- Add parameter `--file` to produce command to allow reading messages from a file directly
- Add parameter `--lineSeparator` to produce command to split multiple messages by a custom delimiter

## 1.6.0 - 2020-02-28

### Added

- Support for Kafka 2.4.0
- Add support for different output format for `describe consumer-groups`, `describe topic` and change default output format.
- Add parameter `--print-configs` to `describe topic` to control if configs shall be printed
- Add parameter `--print-members` to `describe consumer-groups` to control if members shall be printed
- Add parameter `--print-topics` to `describe consumer-groups` to control if topics shall be printed
- Add alias `list` for `get` command (e.g. `list topics`)
- kafkactl is now installable with `homebrew`
- a docker image is now generated with every release

## 1.5.0 - 2020-01-24

### Added
- parameter `--partitioner` for produce command now supports additional partitioners `hash-ref` and `murmur2`. The
 default has also been changed to `murmur2` in order to be consistent with the java kafka client.
- `defaultPartitioner` can now be configured for a context in the config file

## 1.4.0 - 2019-12-13

### Added
- Add rate limiting for multiple messages with `--rate` flag
- A default config file is now generated when no config was found

## 1.3.0 - 2019-11-13

### Added
- Add parameter `--print-headers` to print kafka message headers when consuming messages
- Add parameter `--tail` to consume only the last `n` messages from a topic
- Add parameter `--exit` to stop consuming when the latest offset is reached
- Add parameter `--separator` to consume command to customize the separator
- Allow producing multiple messages from stdin

## 1.2.1 - 2019-05-27

### Fixed
- Fixed producing value as key in case an avro schema registry is configured.

## 1.2.0 - 2019-05-24

### Added
- Additional config file locations added. See README.adoc for details
- Added `offset` parameter to `consume`
- Support for basic auto completion in fish shell
- Add 'config view` command, to view current config
- Add `get consumer-groups` command to list available consumer groups
- Add `describe consumer-group` command to see details of consumer group
- Add `reset offset` command to reset consumer group offsets

### Fixed
- Improved performance of `get topics` through concurrent requests

## 1.1.0 - 2019-03-14

### Added
- Add Avro schema support in `produce` and `consume` commands

## 1.0.0 - 2019-01-23
### Added
- Add `alter topic` command for increasing partition count and editing topic configs
- Add options to read config file from different locations

### Changed
- Sort result of `kafkactl get topics`
- `consume` now uses a simpler consumer without consumerGroup.
- Changed name of global flags to avoid collisions with local flags

### Fixed
- Do not auto-create topics when using `describe`, `consume` or `produce`

## 0.0.1 - 2018-12-12
### Added
- Initial version

[Unreleased]: https://github.com/olivierlacan/keep-a-changelog/compare/0.0.1...HEAD
