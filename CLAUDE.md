# CLAUDE.md - kafkactl Project Guide

## Project Overview

kafkactl is a command-line interface for Apache Kafka written in Go. It supports topic/consumer-group/ACL/broker/user management, message producing/consuming with Avro/Protobuf/JSON schema support, Kubernetes-proxied execution, and an external plugin system for OAuth token providers.

- **Module**: `github.com/deviceinsight/kafkactl/v5`
- **Go version**: 1.24.12
- **Kafka client**: IBM/sarama
- **CLI framework**: spf13/cobra + spf13/viper
- **License**: Apache 2.0

## Build & Run Commands

```bash
# Build the binary
make build

# Run all checks (fmt, lint, cve-check, test, build, docs)
make all

# Format code
make fmt                    # runs gofmt + goimports

# Lint (golangci-lint via Go tool directive)
make lint                   # runs: go tool golangci-lint run

# Unit tests only (skips integration tests)
make test                   # runs: go tool gotestsum --format testname --hide-summary=skipped -- -v -short ./...

# Integration tests (requires Docker)
make integration_test       # starts docker-compose cluster, runs tests, tears down

# CVE check
make cve-check              # runs: go tool govulncheck ./...

# Clean
make clean
```

**Important**: CGO is disabled (`CGO_ENABLED=0`). Tools (golangci-lint, goimports, govulncheck, gotestsum) are managed via `go.mod` `tool` directives, invoked with `go tool <name>`.

## Testing

### Test Naming Convention

- **Unit tests**: Standard Go test functions (e.g., `TestEnvironmentVariableLoading`). No special suffix required.
- **Integration tests**: **Must** be suffixed with `Integration` (e.g., `TestGetTopicsIntegration`). This is enforced by `StartIntegrationTest()` which fatals if the test name doesn't end with `Integration`.

### Running Tests

```bash
# Unit tests only (integration tests are skipped via -short flag)
go test -v -short ./...

# Integration tests (requires Docker Kafka cluster running first)
cd docker && docker compose up -d
go test -v -run Integration ./...

# Stop integration test cluster when done
cd docker && docker compose down
```

### Integration Test Infrastructure

The `docker/docker-compose.yml` provides:
- 3 Kafka brokers (Confluent Platform) with SASL PLAIN, SCRAM-SHA-256, SCRAM-SHA-512
- 1 ZooKeeper
- 1 Schema Registry (Avro + JSON Schema)
- ACL authorizer enabled with `admin` as super user

Test contexts defined in `it-config.yml`:
- `default` - plaintext, ports 19093/29093/39093
- `no-schema-reg` - plaintext, no schema registry
- `sasl-admin` - SASL PLAIN as admin, ports 19092/29092/39092
- `sasl-user` - SASL PLAIN as user
- `k8s-mock` - Kubernetes mock via `docker/kubectl-mock.sh`
- `scram-admin` - SCRAM-SHA-256, ports 19094/29094/39094

### Test Utilities (`internal/testutil/`)

- `StartUnitTest(t)` - call at the start of every unit test
- `StartIntegrationTest(t)` / `StartIntegrationTestWithContext(t, "context")` - call at the start of integration tests
- `CreateKafkaCtlCommand()` - creates a test command instance with captured stdout/stderr
- `CreateTopic(t, prefix, flags...)` - creates a topic with a random suffix to avoid collisions
- `CreateConsumerGroup(t, prefix, topics...)` - creates a consumer group
- `ProduceMessage(t, topic, key, value, partition, offset)` - produces a test message
- `AssertEquals(t, expected, actual)`, `AssertContains`, `AssertErrorContains`, etc. - custom assertions (not using testify)
- `GetPrefixedName(prefix)` - generates `prefix-<random>` names to prevent test interference
- `WithoutBrokerReferences(output)` - normalizes broker IDs/addresses for stable test output

Tests do **not** clean up resources. Random suffixes on topic/group names prevent collisions.

### Test Framework

Uses standard `testing` package with custom assertion helpers in `internal/testutil/`. Does **not** use testify. Tests are run directly with `go test`.

## Project Architecture

### Directory Structure

```
kafkactl/
  main.go                     # Entry point, signal handling, shutdown timeout
  cmd/                        # CLI command definitions (cobra commands)
    root.go                   # Root command, registers all subcommands
    alter/                    # alter topic|partition|broker|user
    attach/                   # attach (K8s interactive shell)
    clone/                    # clone topic|consumer-group
    config/                   # config current-context|get-contexts|use-context|view
    consume/                  # consume <topic>
    create/                   # create topic|consumer-group|acl|user
    deletion/                 # delete topic|consumer-group|offset|records|acl|user
    describe/                 # describe topic|consumer-group|broker|user
    get/                      # get topics|consumer-groups|acl|brokers|users
    produce/                  # produce <topic>
    reset/                    # reset offset
    validation/               # Custom flag validation (at-least-one-required)
    completion.go             # Shell completion command
    version.go                # Version command (Version, GitCommit, BuildTime set via ldflags)
    docs.go                   # Docs generation command
  internal/                   # Internal packages (business logic)
    common-operation.go       # ClientContext creation, sarama client/admin factories, TLS/SASL config
    CachingSchemaRegistry.go  # Schema registry client with caching
    docs-operation.go         # Docs generation logic
    global/                   # Global config initialization, viper setup, env variable aliases
    output/                   # IOStreams, TableWriter, output formatting (json/yaml/table)
    acl/                      # ACL operations
    auth/                     # Token provider loading (plugin + generic script-based)
    broker/                   # Broker operations
    consume/                  # Consumer operations
    consumergroupoffsets/      # Consumer group offset operations
    consumergroups/           # Consumer group operations
    helpers/                  # Shared helpers
      avro/                   # Avro serialization
      protobuf/               # Protobuf serialization
      schemaregistry/         # Schema registry helpers
      scram.go                # SCRAM authentication helper
      TerminalContext.go      # Terminal context for interactive input
    k8s/                      # Kubernetes proxy execution
    partition/                # Partition operations
    producer/                 # Producer operations
    testutil/                 # Test utilities and helpers
    topic/                    # Topic operations
    user/                     # SCRAM user operations
    util/                     # General utility functions
  pkg/                        # Public API packages
    plugins/                  # Plugin system
      plugins.go              # Generic PluginSpec[P, I] type
      auth/                   # Auth plugin interface + RPC implementation
  docker/                     # Docker-compose setup for integration tests
  .github/                    # CI workflows, PR template, contributing guide
```

### Command Pattern

Each command follows this pattern:

1. **Command file** in `cmd/<verb>/` defines the cobra command, flags, and routing
2. **Operation struct** in `internal/<resource>/` implements the business logic
3. **K8s proxy check**: Commands check `internal.IsKubernetesEnabled()` and delegate to `k8s.NewOperation().Run(cmd, args)` if true

Example (`cmd/get/get-topics.go`):
```go
func newGetTopicsCmd() *cobra.Command {
    var flags topic.GetTopicsFlags
    var cmdGetTopics = &cobra.Command{
        Use:   "topics",
        RunE: func(cmd *cobra.Command, args []string) error {
            if internal.IsKubernetesEnabled() {
                return k8s.NewOperation().Run(cmd, args)
            }
            return (&topic.Operation{}).GetTopics(flags)
        },
    }
    // ... flag definitions
}
```

### Key Abstractions

- **`internal.ClientContext`**: Central configuration struct built from viper config. Contains brokers, TLS, SASL, K8s, schema registry, protobuf, producer/consumer settings.
- **`internal.CreateClientContext()`**: Reads the current context from viper and constructs a `ClientContext`.
- **`internal.CreateClient()` / `CreateClusterAdmin()`**: Factory functions that create sarama clients from a `ClientContext`.
- **`output.IOStreams`**: Abstraction over stdin/stdout/stderr/debug. Tests use buffered IOStreams for capturing output.
- **`output.PrintObject(object, format)`**: Renders objects as JSON, YAML, or raw JSON.
- **`output.TableWriter`**: Tab-aligned table output for list commands.

### Kubernetes Proxy

When `kubernetes.enabled` is true in the context config, commands are not executed locally. Instead, kafkactl deploys a pod into the configured K8s cluster and proxies stdin/stdout. Two modes:
- `kafkactl attach` - interactive bash shell in an Ubuntu-based container
- Any other command - runs in a scratch-based container, wires I/O

### Plugin System

Uses HashiCorp `go-plugin` with RPC transport. Currently supports one plugin type:

- **`AccessTokenProvider`** (for OAuth SASL): Interface with `Init(options, brokers)` and `Token()` methods
- Plugin binaries are discovered by name: `kafkactl-<pluginName>-plugin`
- A built-in `generic` token provider executes external scripts for token retrieval

### Configuration

- Config file resolution order: `--config-file` flag > `KAFKA_CTL_CONFIG` env > project-level `kafkactl.yml`/`.kafkactl.yml` > `$HOME/.config/kafkactl/config.yml` > other default locations
- Every config key can be overridden via environment variables (dot to underscore, uppercase). Default context keys can omit the `CONTEXTS_DEFAULT_` prefix.
- Environment variable aliases are defined in `internal/global/env-variables.go`

## CLI Command Hierarchy

```
kafkactl
  alter       topic | partition | broker | user
  attach      (K8s interactive shell)
  clone       topic | consumer-group
  completion  bash | zsh | fish
  config      current-context | get-contexts | use-context | view
  consume     <topic>
  create      topic | consumer-group | acl | user
  delete      topic | consumer-group | offset | records | acl | user
  describe    topic | consumer-group | broker | user
  get         topics | consumer-groups | acl | brokers | users   (alias: list)
  produce     <topic>
  reset       offset
  version
  docs        (hidden, generates command documentation)
```

## Code Style & Conventions

### Linting

Configured in `.golangci.yml` with these linters: `gofmt`, `goimports`, `revive`, `govet`. Timeout: 3 minutes.

Pre-commit hook runs golangci-lint (`.pre-commit-config.yaml`).

### Naming Conventions

- Files: kebab-case (e.g., `common-operation.go`, `get-topics.go`, `k8s-operation.go`)
- Test files: same name with `_test.go` suffix
- Packages: lowercase, short names matching directory
- Command constructors: `NewXxxCmd()` (exported) or `newXxxCmd()` (unexported)
- Operation types: `type Operation struct{}` with methods

### Error Handling

- Uses `github.com/pkg/errors` for wrapping (`errors.Wrap`, `errors.Errorf`)
- Commands return errors via `RunE` (cobra handles display)
- `output.Warnf()` for warnings to stderr
- `output.Debugf()` for debug-level messages (enabled with `-V` flag)

### Output Formats

Commands support `-o` flag with values: `json`, `yaml`, `wide`, `compact`, `none`
- Default: table format via `output.TableWriter`
- Structured: `output.PrintObject(object, format)` for json/yaml

### Commit Message Conventions

From recent history, messages follow these patterns:
- `feat: <description>` - new features
- `fix: <description>` - bug fixes
- `doc: <description>` / `docs: <description>` - documentation changes
- `refactor <description>` - code refactoring
- `chore: <description>` - maintenance
- Release commits: `releases <version>`
- Merge commits: `Merge pull request #N from <branch>`

### Changelog

`CHANGELOG.md` follows [Keep a Changelog](https://keepachangelog.com/) format with `## [Unreleased]` section at the top. PR template requires changes to be documented in the Unreleased section.

## Key Dependencies

| Dependency | Purpose |
|---|---|
| `IBM/sarama` | Apache Kafka client |
| `spf13/cobra` | CLI framework |
| `spf13/viper` | Configuration management |
| `hashicorp/go-plugin` | Plugin system (RPC-based) |
| `linkedin/goavro/v2` | Avro serialization |
| `riferrei/srclient` | Schema Registry client |
| `bufbuild/protocompile` | Protobuf compilation (no protoc needed) |
| `google.golang.org/protobuf` | Protobuf runtime |
| `pkg/errors` | Error wrapping |
| `xdg-go/scram` | SCRAM authentication |
| `go.uber.org/ratelimit` | Producer rate limiting |
| `Rican7/retry` | Retry logic (used in tests) |

## Development Tips

- Integration tests require Docker. Start the cluster with `cd docker && docker compose up -d`, then run tests.
- The `it-config.yml` file configures test contexts. Verify connectivity with: `kafkactl -V -C it-config.yml get brokers`
- When adding a new command, follow the existing pattern: create a command file in `cmd/<verb>/`, implement the operation in `internal/<resource>/`, check for K8s proxy in the `RunE` function.
- All config keys should have corresponding environment variable aliases in `internal/global/env-variables.go`.
