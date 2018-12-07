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

**NOTE:** make sure that `kafkactl` is on PATH otherwise auto-completion won't work.

## configuration

1. create `~/.kafkactl.yml` with a definition of contexts that should be available 

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

2. in order to get auto completion add it in startup script of the shell:

- for `bash` add the following to `~/.bashrc`:
```bash
# kafkactl autocomplete
source <(kafkactl completion bash)
```

- for `zsh` add the following to `~/.zshrc`:
```bash
# kafkactl autocomplete
source <(kafkactl completion zsh)
```

- `fish` is currently not supported. see: https://github.com/spf13/cobra/issues/350