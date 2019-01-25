
# kafkactl

A command-line interface for interaction with Apache Kafka

[![Build Status](https://travis-ci.com/deviceinsight/kafkactl.svg?branch=master)](
  https://travis-ci.com/deviceinsight/kafkactl)

## Features

- Auto-completion
- Configuration of different contexts

## Installation

You can install the pre-compiled binary or compile from source.

### Install the pre-compiled binary

**snapcraft**:

```bash
$ snap install kafkactl
```

**deb/rpm**:

Download the .deb or .rpm from the [releases page](https://github.com/deviceinsight/kafkactl/releases) and install with dpkg -i and rpm -i respectively.

**manually**:

Download the pre-compiled binaries from the [releases page](https://github.com/deviceinsight/kafkactl/releases) and copy to the desired location.

### Compiling from source

```bash
$ go get -u github.com/deviceinsight/kafkactl
```

**NOTE:** make sure that `kafkactl` is on PATH otherwise auto-completion won't work.

## Configuration

### Create a config file

Create `~/.kafkactl/config.yml` with a definition of contexts that should be available

```yaml
contexts:
  localhost:
    brokers:
    - localhost:9092
  remote-cluster:
    brokers:
    - remote-cluster001:9092
    - remote-cluster002:9092
    - remote-cluster003:9092

current-context: localhost
```

The config file location is resolved by
 - checking for a provided commandline argument: `--config-file=$PATH_TO_CONFIG`
 - or by evaluating the environment variable: `export KAFKA_CTL_CONFIG=$PATH_TO_CONFIG`
 - or as default the config file is looked up from one of the following locations: `~/.kafkactl/config.yml`, `/etc/kafkactl/config.yml`


### Auto completion

#### bash

In order to get auto completion add it in startup script of the shell:

- for `bash` add the following to `~/.bashrc`:
```bash
# kafkactl autocomplete
source <(kafkactl completion bash)
```

#### zsh

Create file with completions:

```bash
mkdir ~/.zsh-completions
kafkactl completion zsh > ~/.zsh-completions/_kafkactl
```

To auto-load completion on zsh startup, edit `~/.zshrc`:
```bash
# folder of all of your autocomplete functions
fpath=($HOME/.zsh-completions $fpath)

# enable autocomplete function
autoload -U compinit
compinit
```

#### fish
`fish` is currently not supported. see: https://github.com/spf13/cobra/issues/350

## Examples

### Consuming messages

Consuming messages from a topic can be done with:
```bash
kafkactl consume my-topic
```

In order to consume starting from the oldest offset use:
```bash
kafkactl consume my-topic --from-beginning
```

The following example prints message `key` and `timestamp` as well as `partition` and `offset` in `yaml` format:
```bash
kafkactl consume my-topic --print-keys --print-timestamps -o yaml
```

### Producing messages

Producing messages can be done in multiple ways. If we want to produce a message with `key='my-key'`,
`value='my-value'` to the topic `my-topic` this can be achieved with one of the following commands:

```bash
echo "my-key#my-value" | kafkactl produce my-topic --separator=#
echo "my-value" | kafkactl produce my-topic --key=my-key
kafkactl produce my-topic --key=my-key --value=my-value
```

It is also possible to specify the partition to insert the message:
```bash
kafkactl produce my-topic --key=my-key --value=my-value --partition=2
```

Additionally, a different partitioning scheme can be used. When a `key` is provided the default partitioner
uses the `hash` of the `key` to assign a partition. So the same `key` will end up in the same partition: 
```bash
# the following 3 messages will all be inserted to the same partition
kafkactl produce my-topic --key=my-key --value=my-value
kafkactl produce my-topic --key=my-key --value=my-value
kafkactl produce my-topic --key=my-key --value=my-value

# the following 3 messages will probably be inserted to different partitions
kafkactl produce my-topic --key=my-key --value=my-value --partitioner=random
kafkactl produce my-topic --key=my-key --value=my-value --partitioner=random
kafkactl produce my-topic --key=my-key --value=my-value --partitioner=random
```

### Altering topics

Using the `alter topic` command allows you to change the partition count and topic-level configurations of an existing topic.

The partition count can be increased with:
```bash
kafkactl alter topic my-topic --partitions 32
```

The topic configs can be edited by supplying key value pairs as follows:
```bash
kafkactl alter topic my-topic --config retention.ms=3600 --config cleanup.policy=compact
```
