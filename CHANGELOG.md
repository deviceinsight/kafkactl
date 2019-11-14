# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add rate limiting for multiple messages with `--rate` flag

## 1.3.0 - 2019-11-13

### Added
- Add parameter `--print-headers` to print kafka message headers when consuming messages
- Add parameter `--tail` to consume only the last `n` messages from a topic
- Add parameter `--exit` to stop consuming when the latest offset is reached
- Add parameter `--separator` to consume command to customize the separator
- Allow producing multiple messages from stdin
* Add rate limiting for multiple messages with `--rate` flag

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
