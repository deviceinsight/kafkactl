# kafkactl

command-line interface for interaction with kafka

## features
- auto-completion
- configuration of different contexts

## installation

### from source

```bash
go get -u github.com/random-dwi/kafkactl
```

## configuration

1. create `~/.kafkacli` with a definition of contexts that should be available 

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

2. add the following line to `~/.bashrc` in order to get bash auto completion:
```bash
# kafkactl autocomplete
source <(kafkactl completion bash)
```