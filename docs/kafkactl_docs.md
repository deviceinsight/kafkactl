## kafkactl

command-line interface for Apache Kafka

### Synopsis

A command-line interface the simplifies interaction with Kafka.

### Options

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -h, --help                 help for kafkactl
  -V, --verbose              verbose output
```

### SEE ALSO

* [kafkactl alter](kafkactl_alter.md)	 - alter topics, partitions
* [kafkactl attach](kafkactl_attach.md)	 - run kafkactl pod in kubernetes and attach to it
* [kafkactl completion](kafkactl_completion.md)	 - 
* [kafkactl config](kafkactl_config.md)	 - show and edit configurations
* [kafkactl consume](kafkactl_consume.md)	 - consume messages from a topic
* [kafkactl create](kafkactl_create.md)	 - create topics, consumerGroups, acls
* [kafkactl delete](kafkactl_delete.md)	 - delete topics, consumerGroups, consumer-group-offset, acls
* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups, brokers
* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups, acls, brokers
* [kafkactl produce](kafkactl_produce.md)	 - produce messages to a topic
* [kafkactl reset](kafkactl_reset.md)	 - reset consumerGroupsOffset
* [kafkactl version](kafkactl_version.md)	 - print the version of kafkactl


### kafkactl alter

alter topics, partitions

#### Synopsis

alter topics, partitions

#### Options

```
  -h, --help   help for alter
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl alter partition](kafkactl_alter_partition.md)	 - alter a partition
* [kafkactl alter topic](kafkactl_alter_topic.md)	 - alter a topic


#### kafkactl alter partition

alter a partition

##### Synopsis

alter a partition

```
kafkactl alter partition TOPIC PARTITION [flags]
```

##### Options

```
  -h, --help                  help for partition
  -r, --replicas int32Slice   set replicas for a partition (default [])
  -v, --validate-only         validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](kafkactl_alter.md)	 - alter topics, partitions


#### kafkactl alter topic

alter a topic

##### Synopsis

alter a topic

```
kafkactl alter topic TOPIC [flags]
```

##### Options

```
  -c, --config key=value           configs in format key=value
  -h, --help                       help for topic
  -p, --partitions int32           number of partitions
  -r, --replication-factor int16   replication factor
  -v, --validate-only              validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](#kafkactl-alter)	 - xxxx alter topics, partitions


### kafkactl attach

run kafkactl pod in kubernetes and attach to it

#### Synopsis

run kafkactl pod in kubernetes and attach to it

```
kafkactl attach [flags]
```

#### Options

```
  -h, --help   help for attach
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


### kafkactl completion



#### Synopsis

To load completions:

Bash:

$ source <(kafkactl completion bash)

## To load completions for each session, execute once:
Linux:
  $ kafkactl completion bash > /etc/bash_completion.d/kafkactl
MacOS:
  $ kafkactl completion bash > /usr/local/etc/bash_completion.d/kafkactl

Zsh:

$ source <(kafkactl completion zsh)

## To load completions for each session, execute once:
$ kafkactl completion zsh > "${fpath[1]}/_kafkactl"

Fish:

$ kafkactl completion fish | source

## To load completions for each session, execute once:
$ kafkactl completion fish > ~/.config/fish/completions/kafkactl.fish


```
kafkactl completion [bash|zsh|fish|powershell]
```

#### Options

```
  -h, --help   help for completion
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -e, --exit                                                                             stop consuming when latest offset is reached
  -b, --from-beginning                                                                   set offset for consumer to the oldest offset
  -h, --help                                                                             help for consume
      --key-encoding string                                                              key encoding (auto-detected by default). One of: none|hex|base64
      --key-proto-type string                                                            key protobuf message type
      --offset partition=offset (for partitions not specified, other parameters apply)   offsets in format partition=offset (for partitions not specified, other parameters apply)
  -o, --output string                                                                    output format. One of: json|yaml
  -p, --partitions ints                                                                  partitions to consume. The default is to consume from all partitions.
      --print-headers                                                                    print message headers
  -k, --print-keys                                                                       print message keys
  -a, --print-schema                                                                     print details about avro schema used for decoding
  -t, --print-timestamps                                                                 print message timestamps
      --proto-file strings                                                               additional protobuf description file for searching message description
      --proto-import-path strings                                                        additional path to search files listed in proto 'import' directive
      --protoset-file strings                                                            additional compiled protobuf description file for searching message description
  -s, --separator string                                                                 separator to split key and value (default "#")
      --tail int                                                                         show only the last n messages on the topic (default -1)
      --value-encoding string                                                            value encoding (auto-detected by default). One of: none|hex|base64
      --value-proto-type string                                                          value protobuf message type
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


### kafkactl create

create topics, consumerGroups, acls

#### Synopsis

create topics, consumerGroups, acls

#### Options

```
  -h, --help   help for create
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl create access-control-list](kafkactl_create_access-control-list.md)	 - create an acl
* [kafkactl create consumer-group](kafkactl_create_consumer-group.md)	 - create a consumerGroup
* [kafkactl create topic](kafkactl_create_topic.md)	 - create a topic


#### kafkactl create access-control-list

create an acl

##### Synopsis

create an acl

```
kafkactl create access-control-list [flags]
```

##### Options

```
  -a, --allow                   acl of permissionType 'allow' (choose this or 'deny')
  -c, --cluster                 create acl for the cluster
  -d, --deny                    acl of permissionType 'deny' (choose this or 'allow')
  -g, --group string            create acl for a consumer group
  -h, --help                    help for access-control-list
      --host stringArray        hosts to allow
  -o, --operation stringArray   operations of acl
      --pattern string          pattern type. one of (match, prefixed, literal) (default "literal")
  -p, --principal string        principal to be authenticated
  -t, --topic string            create acl for a topic
  -v, --validate-only           validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](kafkactl_create.md)	 - create topics, consumerGroups, acls


#### kafkactl create consumer-group

create a consumerGroup

##### Synopsis

create a consumerGroup

```
kafkactl create consumer-group GROUP [flags]
```

##### Options

```
  -h, --help              help for consumer-group
      --newest            set the offset to newest offset (for all partitions or the specified partition)
      --offset int        set offset to this value. offset with value -1 is ignored (default -1)
      --oldest            set the offset to oldest offset (for all partitions or the specified partition)
  -p, --partition int32   partition to create group for. -1 stands for all partitions (default -1)
  -t, --topic string      topic to change create group for
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](kafkactl_create.md)	 - create topics, consumerGroups, acls


#### kafkactl create topic

create a topic

##### Synopsis

create a topic

```
kafkactl create topic TOPIC [flags]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](kafkactl_create.md)	 - create topics, consumerGroups, acls


### kafkactl delete

delete topics, consumerGroups, consumer-group-offset, acls

#### Synopsis

delete topics, consumerGroups, consumer-group-offset, acls

#### Options

```
  -h, --help   help for delete
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl delete access-control-list](kafkactl_delete_access-control-list.md)	 - delete an acl
* [kafkactl delete consumer-group](kafkactl_delete_consumer-group.md)	 - delete a consumer-group
* [kafkactl delete consumer-group-offset](kafkactl_delete_consumer-group-offset.md)	 - delete a consumer-group-offset
* [kafkactl delete topic](kafkactl_delete_topic.md)	 - delete a topic


#### kafkactl delete access-control-list

delete an acl

##### Synopsis

delete an acl

```
kafkactl delete access-control-list [flags]
```

##### Options

```
  -a, --allow              acl of permissionType 'allow'
  -c, --cluster            delete acl for the cluster
  -d, --deny               acl of permissionType 'deny'
  -g, --groups             delete acl for a consumer group
  -h, --help               help for access-control-list
  -o, --operation string   operation of acl
      --pattern string     pattern type. one of (any, match, prefixed, literal)
  -t, --topics             delete acl for a topic
  -v, --validate-only      validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](kafkactl_delete.md)	 - delete topics, consumerGroups, consumer-group-offset, acls


#### kafkactl delete consumer-group-offset

delete a consumer-group-offset

##### Synopsis

delete a consumer-group-offset

```
kafkactl delete consumer-group-offset CONSUMER-GROUP --topic=TOPIC --partition=PARTITION [flags]
```

##### Options

```
  -h, --help              help for consumer-group-offset
  -p, --partition int32   delete offset for this partition. -1 stands for all partitions (default -1)
  -t, --topic string      delete offset for this topic
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](kafkactl_delete.md)	 - delete topics, consumerGroups, consumer-group-offset, acls


#### kafkactl delete consumer-group

delete a consumer-group

##### Synopsis

delete a consumer-group

```
kafkactl delete consumer-group CONSUMER-GROUP [flags]
```

##### Options

```
  -h, --help   help for consumer-group
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](kafkactl_delete.md)	 - delete topics, consumerGroups, consumer-group-offset, acls


#### kafkactl delete topic

delete a topic

##### Synopsis

delete a topic

```
kafkactl delete topic TOPIC [flags]
```

##### Options

```
  -h, --help   help for topic
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](kafkactl_delete.md)	 - delete topics, consumerGroups, consumer-group-offset, acls


### kafkactl describe

describe topics, consumerGroups, brokers

#### Synopsis

describe topics, consumerGroups, brokers

#### Options

```
  -h, --help   help for describe
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl describe broker](kafkactl_describe_broker.md)	 - describe a broker
* [kafkactl describe consumer-group](kafkactl_describe_consumer-group.md)	 - describe a consumerGroup
* [kafkactl describe topic](kafkactl_describe_topic.md)	 - describe a topic


#### kafkactl describe broker

describe a broker

##### Synopsis

describe a broker

```
kafkactl describe broker ID [flags]
```

##### Options

```
  -h, --help            help for broker
  -o, --output string   output format. One of: json|yaml|wide
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups, brokers


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
  -o, --output string   output format. One of: json|yaml|wide
  -m, --print-members   print group members (default true)
  -T, --print-topics    print topic details (default true)
  -t, --topic string    show group details for given topic only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups, brokers


#### kafkactl describe topic

describe a topic

##### Synopsis

describe a topic

```
kafkactl describe topic TOPIC [flags]
```

##### Options

```
  -h, --help            help for topic
  -o, --output string   output format. One of: json|yaml|wide
  -c, --print-configs   print configs (default true)
  -s, --skip-empty      show only partitions that have a messages
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](kafkactl_describe.md)	 - describe topics, consumerGroups, brokers


### kafkactl get

get info about topics, consumerGroups, acls, brokers

#### Synopsis

get info about topics, consumerGroups, acls, brokers

#### Options

```
  -h, --help   help for get
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka
* [kafkactl get access-control-list](kafkactl_get_access-control-list.md)	 - list available acls
* [kafkactl get brokers](kafkactl_get_brokers.md)	 - list brokers
* [kafkactl get consumer-groups](kafkactl_get_consumer-groups.md)	 - list available consumerGroups
* [kafkactl get topics](kafkactl_get_topics.md)	 - list available topics


#### kafkactl get access-control-list

list available acls

##### Synopsis

list available acls

```
kafkactl get access-control-list [flags]
```

##### Options

```
  -a, --allow              acl of permissionType 'allow'
  -c, --cluster            list acl for the cluster
  -d, --deny               acl of permissionType 'deny'
  -g, --groups             list acl for consumer groups
  -h, --help               help for access-control-list
      --operation string   operation of acl (default "any")
  -o, --output string      output format. One of: json|yaml
      --pattern string     pattern type. one of (any, match, prefixed, literal) (default "any")
  -t, --topics             list acl for topics
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups, acls, brokers


#### kafkactl get brokers

list brokers

##### Synopsis

list brokers

```
kafkactl get brokers [flags]
```

##### Options

```
  -h, --help            help for brokers
  -o, --output string   output format. One of: json|yaml|compact
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups, acls, brokers


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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups, acls, brokers


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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](kafkactl_get.md)	 - get info about topics, consumerGroups, acls, brokers


### kafkactl produce

produce messages to a topic

#### Synopsis

produce messages to a topic

```
kafkactl produce TOPIC [flags]
```

#### Options

```
  -f, --file string                 file to read input from
  -H, --header key:value            headers in format key:value
  -h, --help                        help for produce
  -k, --key string                  key to use for all messages
      --key-encoding string         key encoding (none by default). One of: none|hex|base64
      --key-proto-type string       key protobuf message type
  -K, --key-schema-version int      avro schema version that should be used for key serialization (default is latest) (default -1)
  -L, --lineSeparator string        separator to split multiple messages from stdin or file (default "\n")
      --max-message-bytes int       the maximum permitted size of a message (defaults to 1000000)
      --null-value                  produce a null value (can be used instead of providing a value with --value)
  -p, --partition int32             partition to produce to (default -1)
  -P, --partitioner murmur2         the partitioning scheme to use. Can be murmur2, `hash`, `hash-ref` `manual`, or `random`. (default is murmur2)
      --proto-file strings          additional protobuf description file for searching message description
      --proto-import-path strings   additional path to search files listed in proto 'import' directive
      --protoset-file strings       additional compiled protobuf description file for searching message description
  -r, --rate int                    amount of messages per second to produce on the topic (default -1)
      --required-acks NoResponse    required acks. One of NoResponse, `WaitForLocal`, `WaitForAll`. (default is WaitForLocal)
  -S, --separator string            separator to split key and value from stdin or file
  -s, --silent                      do not write to standard output
  -v, --value string                value to produce
      --value-encoding string       value encoding (none by default). One of: none|hex|base64
      --value-proto-type string     value protobuf message type
  -i, --value-schema-version int    avro schema version that should be used for value serialization (default is latest) (default -1)
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
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
  -C, --config-file string   config file. one of: [$HOME/.config/kafkactl $HOME/.kafkactl $SNAP_REAL_HOME/.config/kafkactl $SNAP_DATA/kafkactl /etc/kafkactl]
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](kafkactl.md)	 - command-line interface for Apache Kafka


