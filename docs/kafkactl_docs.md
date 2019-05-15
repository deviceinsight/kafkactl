## kafkactl

command-line interface for Apache Kafka

### Synopsis

A command-line interface the simplifies interaction with Kafka.

### Options

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -h, --help                 help for kafkactl
  -V, --verbose              verbose output
```

### SEE ALSO

* [kafkactl alter](kafkactl_alter.md)	 - alter topics
* [kafkactl completion](kafkactl_completion.md)	 - Output shell completion code for the specified shell (bash or zsh)
* [kafkactl config](kafkactl_config.md)	 - show and edit configurations
* [kafkactl consume](kafkactl_consume.md)	 - consume messages from a topic
* [kafkactl create](kafkactl_create.md)	 - create topics
* [kafkactl delete](kafkactl_delete.md)	 - delete topics
* [kafkactl describe](kafkactl_describe.md)	 - describe topics
* [kafkactl get](kafkactl_get.md)	 - get info about topics
* [kafkactl produce](kafkactl_produce.md)	 - produce messages to a topic
* [kafkactl version](kafkactl_version.md)	 - print the version of kafkactl


### kafkactl alter

alter topics

#### Synopsis

alter topics

#### Options

```
  -h, --help   help for alter
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl alter topic](kafkactl_alter_topic.md)	 - alter a topic


#### kafkactl alter topic

alter a topic

##### Synopsis

alter a topic

```
kafkactl alter topic [flags]
```

##### Options

```
  -c, --config key=value   configs in format key=value
  -h, --help               help for topic
  -p, --partitions int32   number of partitions
  -v, --validate-only      validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](kafkactl_alter.md)	 - alter topics


### kafkactl completion

Output shell completion code for the specified shell (bash or zsh)

#### Synopsis

Output shell completion code for the specified shell (bash or zsh)

```
kafkactl completion SHELL
```

#### Options

```
  -h, --help   help for completion
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


### kafkactl config

show and edit configurations

#### Synopsis

show and edit configurations

#### Options

```
  -h, --help   help for config
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl config current-context](kafkactl_config_current-context.md)	 - show current context
* [kafkactl config get-contexts](kafkactl_config_get-contexts.md)	 - list configured contexts
* [kafkactl config use-context](kafkactl_config_use-context.md)	 - switch active context


#### kafkactl config current-context

show current context

##### Synopsis

Displays the name of context that is currently active

```
kafkactl config current-context [flags]
```

##### Options

```
  -h, --help   help for current-context
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](kafkactl_config.md)	 - show and edit configurations


#### kafkactl config get-contexts

list configured contexts

##### Synopsis

Output names of all configured contexts

```
kafkactl config get-contexts [flags]
```

##### Options

```
  -h, --help            help for get-contexts
  -o, --output string   Output format. One of: compact
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](kafkactl_config.md)	 - show and edit configurations


#### kafkactl config use-context

switch active context

##### Synopsis

command to switch active context

```
kafkactl config use-context [flags]
```

##### Options

```
  -h, --help   help for use-context
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](kafkactl_config.md)	 - show and edit configurations


### kafkactl consume

consume messages from a topic

#### Synopsis

consume messages from a topic

```
kafkactl consume TOPIC [flags]
```

#### Options

```
  -b, --from-beginning     set offset for consumer to the oldest offset
  -h, --help               help for consume
  -o, --output string      Output format. One of: json|yaml
  -p, --partitions ints    partitions to consume. The default is to consume from all partitions.
  -k, --print-keys         print message keys
  -a, --print-schema       print details about avro schema used for decoding
  -t, --print-timestamps   print message timestamps
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


### kafkactl create

create topics

#### Synopsis

create topics

#### Options

```
  -h, --help   help for create
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl create topic](kafkactl_create_topic.md)	 - create a topic


#### kafkactl create topic

create a topic

##### Synopsis

create a topic

```
kafkactl create topic [flags]
```

##### Options

```
  -c, --config key=value           configs in format key=value
  -h, --help                       help for topic
  -p, --partitions int32           number of partitions (default 1)
  -r, --replication-factor int16   replication factor (default 1)
  -v, --validate-only              validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](kafkactl_create.md)	 - create topics


### kafkactl delete

delete topics

#### Synopsis

delete topics

#### Options

```
  -h, --help   help for delete
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl delete topic](kafkactl_delete_topic.md)	 - delete a topic


#### kafkactl delete topic

delete a topic

##### Synopsis

delete a topic

```
kafkactl delete topic [flags]
```

##### Options

```
  -h, --help   help for topic
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](kafkactl_delete.md)	 - delete topics


### kafkactl describe

describe topics

#### Synopsis

describe topics

#### Options

```
  -h, --help   help for describe
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl describe topic](kafkactl_describe_topic.md)	 - describe a topic


#### kafkactl describe topic

describe a topic

##### Synopsis

describe a topic

```
kafkactl describe topic [flags]
```

##### Options

```
  -h, --help   help for topic
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](kafkactl_describe.md)	 - describe topics


### kafkactl get

get info about topics

#### Synopsis

get info about topics

#### Options

```
  -h, --help   help for get
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl get topics](kafkactl_get_topics.md)	 - list available topics


#### kafkactl get topics

list available topics

##### Synopsis

list available topics

```
kafkactl get topics [flags]
```

##### Options

```
  -h, --help            help for topics
  -o, --output string   Output format. One of: json|yaml|wide|compact
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics


### kafkactl produce

produce messages to a topic

#### Synopsis

produce messages to a topic

```
kafkactl produce [flags]
```

#### Options

```
  -h, --help                       help for produce
  -k, --key string                 key to use for all messages
  -K, --key-schema-version int     avro schema version that should be used for key serialization (default is latest) (default -1)
  -p, --partition int32            partition to produce to (default -1)
  -P, --partitioner hash           The partitioning scheme to use. Can be hash, `manual`, or `random`
  -S, --separator string           separator to split key and value from stdin
  -s, --silent                     do not write to standard output
  -v, --value string               value to produce
  -i, --value-schema-version int   avro schema version that should be used for value serialization (default is latest) (default -1)
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


### kafkactl version

print the version of kafkactl

#### Synopsis

print the version of kafkactl

```
kafkactl version [flags]
```

#### Options

```
  -h, --help   help for version
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


