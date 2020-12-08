# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
- direct support for kafka clusters [running in kubernetes](https://github.com/deviceinsight/kafkactl/blob/master/README.md#running-in-kubernetes)
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
- Additional config file locations added. See README.md for details
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
