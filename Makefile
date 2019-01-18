BUILD_TS := $(shell date -Iseconds --utc)
COMMIT_SHA := $(shell git rev-parse HEAD)
VERSION := $(shell git describe --abbrev=0 --tags)

.DEFAULT_GOAL := build

export CGO_ENABLED=0
export GOOS=linux
export GO111MODULE=on

project=github.com/deviceinsight/kafkactl
ld_flags := "-X $(project)/cmd.version=$(VERSION) -X $(project)/cmd.gitCommit=$(COMMIT_SHA) -X $(project)/cmd.buildTime=$(BUILD_TS)"

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test: vet
	go test ./...

.PHONY: build
build: test
	go build -ldflags $(ld_flags) -o kafkactl

.PHONY: clean
clean:
	rm -f kafkactl
	go clean -testcache
