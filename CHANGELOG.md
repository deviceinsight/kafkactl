# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add configuration option `tlsInsecure` to be able to connect with TLS to IP address served brokers.

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
