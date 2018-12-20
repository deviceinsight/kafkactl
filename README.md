
# kafkactl - 0.0.1

command-line interface for interaction with kafka

[![Build Status](https://travis-ci.com/deviceinsight/kafkactl.svg?branch=master)](
  https://travis-ci.com/deviceinsight/kafkactl)

## features
- auto-completion
- configuration of different contexts

## installation

### from source

```bash
go get -u github.com/deviceinsight/kafkactl
```

### from binary (linux x64 only)

```bash
## download the release
wget https://github.com/deviceinsight/kafkactl/releases/download/0.0.1/kafkactl_x64.tar.xz
## unpack the binary
tar xvf kafkactl_x64.tar.xz
## remove downloaded file
rm kafkactl_x64.tar.xz
## make kafkactl executable
chmod +x kafkactl
## move the binary in to your PATH.
sudo mv kafkactl /usr/local/bin/kafkactl
```

**NOTE:** make sure that `kafkactl` is on PATH otherwise auto-completion won't work.

## configuration

### create a config file

create `~/.kafkactl.yml` with a definition of contexts that should be available 

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

### auto completion

#### bash

in order to get auto completion add it in startup script of the shell:

- for `bash` add the following to `~/.bashrc`:
```bash
# kafkactl autocomplete
source <(kafkactl completion bash)
```

#### zsh

create file with completions:

```bash
mkdir ~/.zsh-completions
kafkactl completion zsh > ~/.zsh-completions/_kafkactl
```

to auto-load completion on zsh startup, edit `~/.zshrc`:
```bash
# folder of all of your autocomplete functions
fpath=($HOME/.zsh-completions $fpath)

# enable autocomplete function
autoload -U compinit
compinit
```

#### fish
`fish` is currently not supported. see: https://github.com/spf13/cobra/issues/350

## examples

### consuming messages

consuming messages from a topic can be done with:
```bash
kafkactl consume my-topic
```

in order to consume starting from the oldest offset use:
```bash
kafkactl consume my-topic --from-beginning
```

the following example prints message `key` and `timestamp` as well as `partition` and `offset` in `yaml` format:
```bash
kafkactl consume my-topic --print-keys --print-timestamps -o yaml
```

### producing messages

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

Additionally a different partitioning scheme can be used. when a `key` is provided the default partitioner
uses the `hash` of the `key` to assign a partition. so the same `key` will end up in the same partition: 
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

