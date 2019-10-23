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
* [kafkactl completion](kafkactl_completion.md)	 - Output shell completion code for the specified shell (bash,zsh,fish)
* [kafkactl config](kafkactl_config.md)	 - show and edit configurations
* [kafkactl consume](kafkactl_consume.md)	 - consume messages from a topic
* [kafkactl create](kafkactl_create.md)	 - create topics
* [kafkactl delete](kafkactl_delete.md)	 - delete topics
* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups
* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups
* [kafkactl produce](kafkactl_produce.md)	 - produce messages to a topic
* [kafkactl reset](kafkactl_reset.md)	 - reset consumerGroupsOffset
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

Output shell completion code for the specified shell (bash,zsh,fish)

#### Synopsis

Output shell completion code for the specified shell (bash,zsh,fish)

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
* [kafkactl config view](kafkactl_config_view.md)	 - show contents of config file


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
  -o, --output string   output format. One of: compact
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


#### kafkactl config view

show contents of config file

##### Synopsis

Shows the contents of the config file that is currently used

```
kafkactl config view [flags]
```

##### Options

```
  -h, --help   help for view
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
  -b, --from-beginning                                                                   set offset for consumer to the oldest offset
  -h, --help                                                                             help for consume
      --offset partition=offset (for partitions not specified, other parameters apply)   offsets in format partition=offset (for partitions not specified, other parameters apply)
  -o, --output string                                                                    output format. One of: json|yaml
  -p, --partitions ints                                                                  partitions to consume. The default is to consume from all partitions.
      --print-headers                                                                    print message headers
  -k, --print-keys                                                                       print message keys
  -a, --print-schema                                                                     print details about avro schema used for decoding
  -t, --print-timestamps                                                                 print message timestamps
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

describe topics, consumerGroups

#### Synopsis

describe topics, consumerGroups

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
* [kafkactl describe consumer-group](kafkactl_describe_consumer-group.md)	 - describe a consumerGroup
* [kafkactl describe topic](kafkactl_describe_topic.md)	 - describe a topic


#### kafkactl describe consumer-group

describe a consumerGroup

##### Synopsis

describe a consumerGroup

```
kafkactl describe consumer-group GROUP [flags]
```

##### Options

```
  -h, --help            help for consumer-group
  -l, --only-with-lag   show only partitions that have a lag
  -t, --topic string    show group details for given topic only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups


#### kafkactl describe topic

describe a topic

##### Synopsis

describe a topic

```
kafkactl describe topic TOPIC [flags]
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

* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups


### kafkactl get

get info about topics, consumerGroups

#### Synopsis

get info about topics, consumerGroups

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
* [kafkactl get consumer-groups](kafkactl_get_consumer-groups.md)	 - list available consumerGroups
* [kafkactl get topics](kafkactl_get_topics.md)	 - list available topics


#### kafkactl get consumer-groups

list available consumerGroups

##### Synopsis

list available consumerGroups

```
kafkactl get consumer-groups [flags]
```

##### Options

```
  -h, --help            help for consumer-groups
  -o, --output string   output format. One of: json|yaml|wide|compact
  -t, --topic string    show groups for given topic only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups


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
  -o, --output string   output format. One of: json|yaml|wide|compact
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups


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
  -P, --partitioner hash           the partitioning scheme to use. Can be hash, `manual`, or `random`
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


### kafkactl reset

reset consumerGroupsOffset

#### Synopsis

reset consumerGroupsOffset

#### Options

```
  -h, --help   help for reset
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl reset consumer-group-offset](kafkactl_reset_consumer-group-offset.md)	 - reset a consumer group offset


#### kafkactl reset consumer-group-offset

reset a consumer group offset

##### Synopsis

reset a consumer group offset

```
kafkactl reset consumer-group-offset GROUP [flags]
```

##### Options

```
  -e, --execute           execute the reset (as default only the results are displayed for validation)
  -h, --help              help for consumer-group-offset
      --newest            set the offset to newest offset (for all partitions or the specified partition)
      --offset int        set offset to this value. offset with value -1 is ignored (default -1)
      --oldest            set the offset to oldest offset (for all partitions or the specified partition)
  -o, --output string     output format. One of: json|yaml
  -p, --partition int32   partition to apply the offset. -1 stands for all partitions (default -1)
  -t, --topic string      topic to change offset for
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl reset](kafkactl_reset.md)	 - reset consumerGroupsOffset


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


