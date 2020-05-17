BUILD_TS := $(shell date -Iseconds --utc)
COMMIT_SHA := $(shell git rev-parse HEAD)
VERSION := $(shell git describe --abbrev=0 --tags)

.DEFAULT_GOAL := build

export CGO_ENABLED=0
export GOOS=linux
export GO111MODULE=on

project=github.com/deviceinsight/kafkactl
ld_flags := "-X $(project)/cmd.version=$(VERSION) -X $(project)/cmd.gitCommit=$(COMMIT_SHA) -X $(project)/cmd.buildTime=$(BUILD_TS)"

FILES    := $(shell find . -name '*.go' -type f -not -name '*.pb.go' -not -name '*_generated.go' -not -name '*_test.go')
TESTS    := $(shell find . -name '*.go' -type f -not -name '*.pb.go' -not -name '*_generated.go' -name '*_test.go')

.DEFAULT_GOAL := all
.PHONY: all
all: vet fmt lint test build docs

.PHONY: vet
vet:
	go vet ./...

fmt:
	gofmt -s -l -w $(FILES) $(TESTS)

lint:
	golangci-lint run

.PHONY: test
test:
	go test -v -short ./...

.PHONY: pre_integration_test
pre_integration_test:
	docker-compose -f docker/docker-compose.yml up -d

.PHONY: integration_test
integration_test:
	go test -run Integration ./...

.PHONY: post_integration_test
post_integration_test:
	docker-compose -f docker/docker-compose.yml down

.PHONY: build
build:
	go build -ldflags $(ld_flags) -o kafkactl

.PHONY: docs
docs: build
	touch /tmp/empty.yaml
	./kafkactl docs --directory docs --single-page --config-file=/tmp/empty.yaml

.PHONY: clean
clean:
	rm -f kafkactl
	go clean -testcache

# usage make version=0.0.4 release
#
# manually executing goreleaser:
# export GITHUB_TOKEN=xyz
# snapcraft login
# goreleaser --rm-dist (--skip-validate)
#
.PHONY: release
release:
	current_date=`date "+%Y-%m-%d"`; eval "sed -i 's/## \[Unreleased\].*/## [Unreleased]\n\n## $$version - $$current_date/g' CHANGELOG.md"
	git add "CHANGELOG.md"
	git commit -m "releases $(version)"
	git tag -a $(version) -m "release $(version)"
	git push origin
	git push origin $(version)
