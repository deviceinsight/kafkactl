
# kafkactl

A command-line interface for interaction with Apache Kafka

[![Build Status](https://travis-ci.com/deviceinsight/kafkactl.svg?branch=master)](https://travis-ci.com/deviceinsight/kafkactl)
| [![command docs](https://img.shields.io/badge/command-docs-blue.svg)](https://deviceinsight.github.io/kafkactl/)  

## Features

- command auto-completion for bash, zsh, fish
- support for avro schemas
- Configuration of different contexts

[![asciicast](https://asciinema.org/a/vmxrTA0h8CAXPnJnSFk5uHKzr.svg)](https://asciinema.org/a/vmxrTA0h8CAXPnJnSFk5uHKzr)

## Installation

You can install the pre-compiled binary or compile from source.

### Install the pre-compiled binary

**snapcraft**:

```bash
snap install kafkactl
```

**homebrew**:
```bash
# install tap repostory once
brew tap deviceinsight/packages
# install kafkactl
brew install deviceinsight/packages/kafkactl
# upgrade kafkactl
brew upgrade deviceinsight/packages/kafkactl
```

**deb/rpm**:

Download the .deb or .rpm from the [releases page](https://github.com/deviceinsight/kafkactl/releases) and install with dpkg -i and rpm -i respectively.

**manually**:

Download the pre-compiled binaries from the [releases page](https://github.com/deviceinsight/kafkactl/releases) and copy to the desired location.

### Compiling from source

```bash
go get -u github.com/deviceinsight/kafkactl
```

**NOTE:** make sure that `kafkactl` is on PATH otherwise auto-completion won't work.

## Configuration

If no config file is found, a default config is generated in `$HOME/.config/kafkactl/config.yml`.
This configuration is suitable to get started with a single node cluster on a local machine. 

### Create a config file

Create `$HOME/.config/kafkactl/config.yml` with a definition of contexts that should be available

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

    # optional: tls config
    tlsCA: my-ca
    tlsCert: my-cert
    tlsCertKey: my-key
    # set tlsInsecure to true to ignore all tls verification (defaults to false)
    tlsInsecure: false
    
    # optional: clientID config (defaults to kafkactl-{username})
    clientID: my-client-id
    
    # optional: kafkaVersion (defaults to 2.0.0)
    kafkaVersion: 1.1.1

    # optional: avro schema registry
    avro:
      schemaRegistry: localhost:8081
    
    # optional: changes the default partitioner
    defaultPartitioner: "hash"


current-context: localhost
```

The config file location is resolved by
 * checking for a provided commandline argument: `--config-file=$PATH_TO_CONFIG`
 * or by evaluating the environment variable: `export KAFKA_CTL_CONFIG=$PATH_TO_CONFIG`
 * or as default the config file is looked up from one of the following locations:
   * `$HOME/.config/kafkactl/config.yml`
   * `$HOME/.kafkactl/config.yml`
   * `$SNAP_DATA/kafkactl/config.yml`
   * `/etc/kafkactl/config.yml`

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

Create file with completions:

```bash
mkdir ~/.config/fish/completions
kafkactl completion fish > ~/.config/fish/completions/kafkactl.fish
```

## Running in docker

Assuming your Kafka broker is accessible as `kafka:9092`, you can list topics by running: 

```bash
docker run --env BROKER=kafka:9092 deviceinsight/kafkactl:latest get topics
```

If a more elaborate config is needed, you can mount it as a volume:

```bash
docker run -v /absolute/path/to/config.yml:/etc/kafkactl/config.yml deviceinsight/kafkactl get topics
``` 

## Command documentation

The documentation for all available commands can be found here:

[![command docs](https://img.shields.io/badge/command-docs-blue.svg)](https://deviceinsight.github.io/kafkactl/)


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

Headers of kafka messages can be printed with the parameter `--print-headers` e.g.:
```bash
kafkactl consume my-topic --print-headers -o yaml
```

If one is only interested in the last `n` messages this can be achieved by `--tail` e.g.:
```bash
kafkactl consume my-topic --tail=5
```

The consumer can be stopped when the latest offset is reached using `--exit` parameter e.g.:
```bash
kafkactl consume my-topic --from-beginning --exit
```

### Producing messages

Producing messages can be done in multiple ways. If we want to produce a message with `key='my-key'`,
`value='my-value'` to the topic `my-topic` this can be achieved with one of the following commands:

```bash
echo "my-key#my-value" | kafkactl produce my-topic --separator=#
echo "my-value" | kafkactl produce my-topic --key=my-key
kafkactl produce my-topic --key=my-key --value=my-value
```

If we have a file containing messages where each line contains `key` and `value` separated by `#`, the file can be
used as input to produce messages to topic `my-topic`:

```bash
cat myfile | kafkactl produce my-topic --separator=#
```

**NOTE:** if the file was generated with `kafkactl consume --print-keys --print-timestamps my-topic` the produce
command is able to detect the message timestamp in the input and will ignore it. 

the number of messages produced per second can be controlled with the `--rate` paramter:

```bash
cat myfile | kafkactl produce my-topic --separator=# --rate=200
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

### Avro support

In order to enable avro support you just have to add the schema registry to your configuration:
```$yaml
contexts:
  localhost:
    avro:
      schemaRegistry: localhost:8081
```

#### Producing to an avro topic

`kafkactl` will lookup the topic in the schema registry in order to determine if key or value needs to be avro encoded.
If producing with the latest `schemaVersion` is sufficient, no additional configuration is needed an `kafkactl` handles
this automatically.

If however one needs to produce an older `schemaVersion` this can be achieved by providing the parameters `keySchemaVersion`, `valueSchemaVersion`.

##### Example

```bash
# create a topic
kafkactl create topic avro_topic
# add a schema for the topic value
curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
--data '{"schema": "{\"type\": \"record\", \"name\": \"LongList\", \"fields\" : [{\"name\": \"next\", \"type\": [\"null\", \"LongList\"], \"default\": null}]}"}' \
http://localhost:8081/subjects/avro_topic-value/versions
# produce a message
kafkactl produce avro_topic --value {\"next\":{\"LongList\":{}}}
# consume the message
kafkactl consume avro_topic --from-beginning --print-schema -o yaml
```

#### Consuming from an avro topic

As for producing `kafkactl` will also lookup the topic in the schema registry to determine if key or value needs to be
decoded with an avro schema.

The `consume` command handles this automatically and no configuration is needed.

An additional parameter `print-schema` can be provided to display the schema used for decoding.

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

### Consumer groups

In order to get a list of consumer groups the `get consumer-groups` command can be used:
```bash
# all available consumer groups
kafkactl get consumer-groups 
# only consumer groups for a single topic
kafkactl get consumer-groups --topic my-topic
# using command alias
kafkactl get cg
```

To get detailed information about the consumer group use `describe consumer-group`. If the parameter `--partitions`
is provided details will be printed for each partition otherwise the partitions are aggregated to the clients.

```bash
# describe a consumer group
kafkactl describe consumer-group my-group 
# show partition details only for partitions with lag
kafkactl describe consumer-group my-group --only-with-lag
# show details only for a single topic
kafkactl describe consumer-group my-group --topic my-topic
# using command alias
kafkactl describe cg my-group
```

### Reset consumer group offsets

in order to ensure the reset does what it is expected, per default only
the results are printed without actually executing it. Use the additional parameter `--execute` to perform the reset. 

```bash
# reset offset of for all partitions to oldest offset
kafkactl reset offset my-group --topic my-topic --oldest
# reset offset of for all partitions to newest offset
kafkactl reset offset my-group --topic my-topic --newest
# reset offset for a single partition to specific offset
kafkactl reset offset my-group --topic my-topic --partition 5 --offset 100
```
