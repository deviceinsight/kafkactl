## kafkactl

command-line interface for Apache Kafka

### Synopsis

A command-line interface the simplifies interaction with Kafka.

### Options

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -h, --help                 help for kafkactl
  -V, --verbose              verbose output
```

### SEE ALSO

* [kafkactl alter](#kafkactl-alter)	 - alter topics, partitions, brokers, users
* [kafkactl attach](#kafkactl-attach)	 - run kafkactl pod in kubernetes and attach to it
* [kafkactl clone](#kafkactl-clone)	 - clone topics, consumerGroups
* [kafkactl completion](#kafkactl-completion)	 - generate shell auto-completion file
* [kafkactl config](#kafkactl-config)	 - show and edit configurations
* [kafkactl consume](#kafkactl-consume)	 - consume messages from a topic
* [kafkactl create](#kafkactl-create)	 - create topics, consumerGroups, acls, users
* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users
* [kafkactl describe](#kafkactl-describe)	 - describe topics, consumerGroups, brokers, users
* [kafkactl get](#kafkactl-get)	 - get info about topics, consumerGroups, acls, brokers, users
* [kafkactl produce](#kafkactl-produce)	 - produce messages to a topic
* [kafkactl reset](#kafkactl-reset)	 - reset consumerGroupsOffset
* [kafkactl version](#kafkactl-version)	 - print the version of kafkactl


### kafkactl alter

alter topics, partitions, brokers, users

#### Options

```
  -h, --help   help for alter
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl alter broker](#kafkactl-alter-broker)	 - alter a broker
* [kafkactl alter partition](#kafkactl-alter-partition)	 - alter a partition
* [kafkactl alter topic](#kafkactl-alter-topic)	 - alter a topic
* [kafkactl alter user](#kafkactl-alter-user)	 - alter a SCRAM user


#### kafkactl alter broker

alter a broker

```
kafkactl alter broker BROKER [flags]
```

##### Options

```
  -c, --config key=value   configs in format key=value
  -h, --help               help for broker
  -v, --validate-only      validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](#kafkactl-alter)	 - alter topics, partitions, brokers, users


#### kafkactl alter partition

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](#kafkactl-alter)	 - alter topics, partitions, brokers, users


#### kafkactl alter topic

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](#kafkactl-alter)	 - alter topics, partitions, brokers, users


#### kafkactl alter user

alter a SCRAM user

```
kafkactl alter user USERNAME [flags]
```

##### Options

```
  -h, --help               help for user
  -i, --iterations int32   SCRAM iterations (default 4096)
  -m, --mechanism string   SCRAM mechanism (SCRAM-SHA-256, SCRAM-SHA-512) (default "SCRAM-SHA-256")
  -p, --password string    new user password
  -s, --salt string        custom salt (base64 encoded, generated if not provided)
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl alter](#kafkactl-alter)	 - alter topics, partitions, brokers, users


### kafkactl attach

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka


### kafkactl clone

clone topics, consumerGroups

#### Options

```
  -h, --help   help for clone
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl clone consumer-group](#kafkactl-clone-consumer-group)	 - clone existing consumerGroup with all offsets
* [kafkactl clone topic](#kafkactl-clone-topic)	 - clone existing topic (number of partitions, replication factor, config entries) to new one


#### kafkactl clone consumer-group

clone existing consumerGroup with all offsets

```
kafkactl clone consumer-group SOURCE_GROUP TARGET_GROUP [flags]
```

##### Options

```
  -h, --help   help for consumer-group
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl clone](#kafkactl-clone)	 - clone topics, consumerGroups


#### kafkactl clone topic

clone existing topic (number of partitions, replication factor, config entries) to new one

```
kafkactl clone topic SOURCE_TOPIC TARGET_TOPIC [flags]
```

##### Options

```
  -h, --help   help for topic
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl clone](#kafkactl-clone)	 - clone topics, consumerGroups


### kafkactl completion

generate shell auto-completion file

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka


### kafkactl config

show and edit configurations

#### Options

```
  -h, --help   help for config
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl config current-context](#kafkactl-config-current-context)	 - show current context
* [kafkactl config get-contexts](#kafkactl-config-get-contexts)	 - list configured contexts
* [kafkactl config use-context](#kafkactl-config-use-context)	 - switch active context
* [kafkactl config view](#kafkactl-config-view)	 - show current config


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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](#kafkactl-config)	 - show and edit configurations


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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](#kafkactl-config)	 - show and edit configurations


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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](#kafkactl-config)	 - show and edit configurations


#### kafkactl config view

show current config

##### Synopsis

Shows the merged config that is currently used

```
kafkactl config view [flags]
```

##### Options

```
  -h, --help   help for view
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl config](#kafkactl-config)	 - show and edit configurations


### kafkactl consume

consume messages from a topic

```
kafkactl consume TOPIC [flags]
```

#### Options

```
  -e, --exit                                                                             stop consuming when latest offset is reached
  -b, --from-beginning                                                                   set offset for consumer to the oldest offset
      --from-timestamp string                                                            consume data from offset of given timestamp
  -g, --group string                                                                     consumer group to join
  -h, --help                                                                             help for consume
  -i, --isolation-level string                                                           isolationLevel to use. One of: ReadUncommitted|ReadCommitted
      --key-encoding string                                                              key encoding (auto-detected by default). One of: none|hex|base64
      --key-proto-type string                                                            key protobuf message type
      --max-messages int                                                                 stop consuming after n messages have been read (default -1)
      --offset partition=offset (for partitions not specified, other parameters apply)   offsets in format partition=offset (for partitions not specified, other parameters apply)
  -o, --output string                                                                    output format. One of: json|yaml
  -p, --partitions ints                                                                  partitions to consume. The default is to consume from all partitions.
      --print-headers                                                                    print message headers
  -k, --print-keys                                                                       print message keys
      --print-partitions                                                                 print message partitions
  -a, --print-schema                                                                     print details about schema used for decoding
  -t, --print-timestamps                                                                 print message timestamps
      --proto-file strings                                                               additional protobuf description file for searching message description
      --proto-import-path strings                                                        additional path to search files listed in proto 'import' directive
      --proto-marshal-option strings                                                     json marshall options to use for protobuf. Format is key=value. Valid keys are allowpartial,useprotonames,useenumnumbers,emitdefaultvalues,emitdefaultvalues
      --protoset-file strings                                                            additional compiled protobuf description file for searching message description
  -s, --separator string                                                                 separator to split key and value (default "#")
      --tail int                                                                         show only the last n messages on the topic (default -1)
      --to-timestamp string                                                              consume data till offset of given timestamp
      --value-encoding string                                                            value encoding (auto-detected by default). One of: none|hex|base64
      --value-proto-type string                                                          value protobuf message type
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka


### kafkactl create

create topics, consumerGroups, acls, users

#### Options

```
  -h, --help   help for create
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl create access-control-list](#kafkactl-create-access-control-list)	 - create an acl
* [kafkactl create consumer-group](#kafkactl-create-consumer-group)	 - create a consumerGroup
* [kafkactl create topic](#kafkactl-create-topic)	 - create a topic
* [kafkactl create user](#kafkactl-create-user)	 - create a SCRAM user


#### kafkactl create access-control-list

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](#kafkactl-create)	 - create topics, consumerGroups, acls, users


#### kafkactl create consumer-group

create a consumerGroup

```
kafkactl create consumer-group GROUP [flags]
```

##### Options

```
  -h, --help                help for consumer-group
      --newest              set the offset to newest offset (for all partitions or the specified partition)
      --offset int          set offset to this value. offset with value -1 is ignored (default -1)
      --oldest              set the offset to oldest offset (for all partitions or the specified partition)
  -p, --partition int32     partition to create group for. -1 stands for all partitions (default -1)
  -t, --topic stringArray   one or more topics to create group for
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](#kafkactl-create)	 - create topics, consumerGroups, acls, users


#### kafkactl create topic

create a topic

```
kafkactl create topic TOPIC [flags]
```

##### Options

```
  -c, --config key=value           configs in format key=value
  -f, --file string                file with topic description
  -h, --help                       help for topic
  -p, --partitions int32           number of partitions (default 1)
  -r, --replication-factor int16   replication factor (default -1)
  -v, --validate-only              validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](#kafkactl-create)	 - create topics, consumerGroups, acls, users


#### kafkactl create user

create a SCRAM user

```
kafkactl create user USERNAME [flags]
```

##### Options

```
  -h, --help               help for user
  -i, --iterations int32   SCRAM iterations (default 4096)
  -m, --mechanism string   SCRAM mechanism (SCRAM-SHA-256, SCRAM-SHA-512) (default "SCRAM-SHA-256")
  -p, --password string    user password
  -s, --salt string        custom salt (base64 encoded, generated if not provided)
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl create](#kafkactl-create)	 - create topics, consumerGroups, acls, users


### kafkactl delete

delete topics, consumerGroups, consumer-group-offset, acls, records, users

#### Options

```
  -h, --help   help for delete
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl delete access-control-list](#kafkactl-delete-access-control-list)	 - delete an acl
* [kafkactl delete consumer-group](#kafkactl-delete-consumer-group)	 - delete a consumer-group
* [kafkactl delete consumer-group-offset](#kafkactl-delete-consumer-group-offset)	 - delete a consumer-group-offset
* [kafkactl delete records](#kafkactl-delete-records)	 - delete a records from a topic
* [kafkactl delete topic](#kafkactl-delete-topic)	 - delete a topic
* [kafkactl delete user](#kafkactl-delete-user)	 - delete SCRAM user credentials


#### kafkactl delete access-control-list

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
      --host string        host of acl
  -o, --operation string   operation of acl
      --pattern string     pattern type. one of (any, match, prefixed, literal)
  -p, --principal string   principal of acl
  -t, --topics             delete acl for a topic
  -v, --validate-only      validate only
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users


#### kafkactl delete consumer-group-offset

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users


#### kafkactl delete consumer-group

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users


#### kafkactl delete records

delete a records from a topic

```
kafkactl delete records TOPIC [flags]
```

##### Options

```
  -h, --help                      help for records
      --offset partition=offset   offsets in format partition=offset. records with smaller offset will be deleted.
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users


#### kafkactl delete topic

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users


#### kafkactl delete user

delete SCRAM user credentials

```
kafkactl delete user USERNAME [flags]
```

##### Options

```
  -h, --help               help for user
  -m, --mechanism string   SCRAM mechanism to delete (SCRAM-SHA-256, SCRAM-SHA-512) (default "SCRAM-SHA-256")
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl delete](#kafkactl-delete)	 - delete topics, consumerGroups, consumer-group-offset, acls, records, users


### kafkactl describe

describe topics, consumerGroups, brokers, users

#### Options

```
  -h, --help   help for describe
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl describe broker](#kafkactl-describe-broker)	 - describe a broker
* [kafkactl describe consumer-group](#kafkactl-describe-consumer-group)	 - describe a consumerGroup
* [kafkactl describe topic](#kafkactl-describe-topic)	 - describe a topic
* [kafkactl describe user](#kafkactl-describe-user)	 - describe a SCRAM user


#### kafkactl describe broker

describe a broker

```
kafkactl describe broker ID [flags]
```

##### Options

```
  -a, --all-configs     print all configs including defaults
  -h, --help            help for broker
  -o, --output string   output format. One of: json|yaml|wide
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](#kafkactl-describe)	 - describe topics, consumerGroups, brokers, users


#### kafkactl describe consumer-group

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](#kafkactl-describe)	 - describe topics, consumerGroups, brokers, users


#### kafkactl describe topic

describe a topic

```
kafkactl describe topic TOPIC [flags]
```

##### Options

```
  -a, --all-configs     print all configs including defaults
  -h, --help            help for topic
  -o, --output string   output format. One of: json|yaml|wide
  -s, --skip-empty      show only partitions that have a messages
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](#kafkactl-describe)	 - describe topics, consumerGroups, brokers, users


#### kafkactl describe user

describe a SCRAM user

```
kafkactl describe user USERNAME [flags]
```

##### Options

```
  -h, --help            help for user
  -o, --output string   output format. One of: json|yaml
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl describe](#kafkactl-describe)	 - describe topics, consumerGroups, brokers, users


### kafkactl get

get info about topics, consumerGroups, acls, brokers, users

#### Options

```
  -h, --help   help for get
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl get access-control-list](#kafkactl-get-access-control-list)	 - list available acls
* [kafkactl get brokers](#kafkactl-get-brokers)	 - list brokers
* [kafkactl get consumer-groups](#kafkactl-get-consumer-groups)	 - list available consumerGroups
* [kafkactl get topics](#kafkactl-get-topics)	 - list available topics
* [kafkactl get users](#kafkactl-get-users)	 - get SCRAM users


#### kafkactl get access-control-list

list available acls

```
kafkactl get access-control-list [flags]
```

##### Options

```
  -a, --allow                  acl of permissionType 'allow'
  -c, --cluster                list acl for the cluster
  -d, --deny                   acl of permissionType 'deny'
  -g, --groups                 list acl for consumer groups
  -h, --help                   help for access-control-list
      --host string            host of acl
      --operation string       operation of acl (default "any")
  -o, --output string          output format. One of: json|yaml
      --pattern string         pattern type. one of (any, match, prefixed, literal) (default "any")
  -p, --principal string       principal of acl
  -r, --resource-name string   resource name of acl (e.g. topic name)
  -t, --topics                 list acl for topics
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](#kafkactl-get)	 - get info about topics, consumerGroups, acls, brokers, users


#### kafkactl get brokers

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](#kafkactl-get)	 - get info about topics, consumerGroups, acls, brokers, users


#### kafkactl get consumer-groups

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](#kafkactl-get)	 - get info about topics, consumerGroups, acls, brokers, users


#### kafkactl get topics

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](#kafkactl-get)	 - get info about topics, consumerGroups, acls, brokers, users


#### kafkactl get users

get SCRAM users

```
kafkactl get users [flags]
```

##### Options

```
  -h, --help            help for users
  -o, --output string   output format. One of: json|yaml
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl get](#kafkactl-get)	 - get info about topics, consumerGroups, acls, brokers, users


### kafkactl produce

produce messages to a topic

```
kafkactl produce TOPIC [flags]
```

#### Options

```
  -f, --file string                 file to read input from
  -H, --header key:value            headers in format key:value
  -h, --help                        help for produce
      --input-format string         input format. One of: csv,json (default is csv)
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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka


### kafkactl reset

reset consumerGroupsOffset

#### Options

```
  -h, --help   help for reset
```

#### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka
* [kafkactl reset consumer-group-offset](#kafkactl-reset-consumer-group-offset)	 - reset a consumer group offset


#### kafkactl reset consumer-group-offset

reset a consumer group offset

```
kafkactl reset consumer-group-offset GROUP [flags]
```

##### Options

```
      --all-topics           do the operation for all topics in the consumer group
  -e, --execute              execute the reset (as default only the results are displayed for validation)
  -h, --help                 help for consumer-group-offset
      --newest               set the offset to newest offset (for all partitions or the specified partition)
      --offset int           set offset to this value. offset with value -1 is ignored (default -1)
      --oldest               set the offset to oldest offset (for all partitions or the specified partition)
  -o, --output string        output format. One of: json|yaml
  -p, --partition int32      partition to apply the offset. -1 stands for all partitions (default -1)
      --to-datetime string   set the offset to offset of given timestamp
  -t, --topic stringArray    one ore more topics to change offset for
```

##### Options inherited from parent commands

```
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

##### SEE ALSO

* [kafkactl reset](#kafkactl-reset)	 - reset consumerGroupsOffset


### kafkactl version

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
  -C, --config-file string   config file. default locations: [$HOME/.config/kafkactl $HOME/.kafkactl $APPDATA/kafkactl /etc/kafkactl]
      --context string       The name of the context to use
  -V, --verbose              verbose output
```

#### SEE ALSO

* [kafkactl](#kafkactl)	 - command-line interface for Apache Kafka


