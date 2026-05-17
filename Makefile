all: check test

ifndef CASSANDRA_IMAGE
CASSANDRA_IMAGE := cassandra:5.0
endif

ifndef GOTEST_CPU
GOTEST_CPU := 1
endif

ifndef GOPATH
GOPATH := $(shell go env GOPATH)
endif

ifndef GOBIN
GOBIN := $(GOPATH)/bin
endif
export PATH := $(GOBIN):$(PATH)

GOOS := $(shell uname | tr '[:upper:]' '[:lower:]')
GOARCH := $(shell go env GOARCH)

GOLANGCI_VERSION := 2.5.0

ifeq ($(GOARCH),arm64)
	GOLANGCI_DOWNLOAD_URL := "https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_VERSION)/golangci-lint-$(GOLANGCI_VERSION)-$(GOOS)-arm64.tar.gz"
else ifeq ($(GOARCH),amd64)
	GOLANGCI_DOWNLOAD_URL := "https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_VERSION)/golangci-lint-$(GOLANGCI_VERSION)-$(GOOS)-amd64.tar.gz"
else
	@printf 'Unknown architecture "%s"\n', "$(GOARCH)"
	@exit 69
endif

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: check
check: .require-golangci-lint
	@golangci-lint run ./...

.PHONY: fix
fix: .require-golangci-lint .require-fieldalignment
	@$(MAKE) fmt
	@golangci-lint run --fix ./...
	@fieldalignment -test=false -fix  ./...

GOTEST := go test -cpu $(GOTEST_CPU) -count=1 -cover -race -tags all

.PHONY: test
test: start-cassandra
	echo "==> Running tests..."
	echo "==> Running tests... in ."
	@$(GOTEST) .
	echo "==> Running tests... in ./qb"
	@$(GOTEST) ./qb
	echo "==> Running tests... in ./table"
	@$(GOTEST) ./table
	echo "==> Running tests... in ./dbutil"
	@$(GOTEST) ./dbutil
	echo "==> Running tests... in ./cmd/schemagen"
	@$(GOTEST) ./cmd/schemagen
	echo "==> Running tests... in ./cmd/schemagen/testdata"
	@cd ./cmd/schemagen/testdata ; go mod tidy ; $(GOTEST) .; cd ../../..
	echo "==> Running tests... in ./examples/..."
	@$(GOTEST) ./examples/...

.PHONY: test-unit
test-unit:
	@echo "==> Running unit tests (no docker, no integration tag)..."
	@go test -count=1 -cover -race ./...

.PHONY: test-coverage
test-coverage: start-cassandra
	@echo "==> Running tests with coverage profile..."
	# -p 1: cqlxtest shares one keyspace across all packages and races if run
	# in parallel. See cqlxtest/cqlxtest.go's initOnce comment.
	@go test -cpu $(GOTEST_CPU) -p 1 -count=1 -race -tags all \
		-coverprofile=coverage.out -covermode=atomic -coverpkg=./... \
		./...

.PHONY: bench
bench:
	@go test -cpu $(GOTEST_CPU) -tags all -run=XXX -bench=. -benchmem ./...

.PHONY: run-examples
run-examples:
	@go test -tags all -v -run=Example ./...

.PHONY: start-cassandra
start-cassandra:
	@if bash -c '</dev/tcp/127.0.0.1/9042' 2>/dev/null; then \
		echo "==> Cassandra on 127.0.0.1:9042 already reachable, reusing"; \
	else \
		echo "==> Running test instance of Cassandra $(CASSANDRA_IMAGE)"; \
		docker rm -f cqlx-cassandra >/dev/null 2>&1 || true; \
		docker pull $(CASSANDRA_IMAGE); \
		docker run --name cqlx-cassandra -p 9042:9042 --rm -d \
			--entrypoint bash $(CASSANDRA_IMAGE) \
			-c "sed -i 's/^materialized_views_enabled:.*/materialized_views_enabled: true/' /etc/cassandra/cassandra.yaml && exec docker-entrypoint.sh cassandra -f"; \
		until docker exec cqlx-cassandra cqlsh -e "DESCRIBE KEYSPACES" >/dev/null 2>&1; do sleep 2; done; \
	fi

.PHONY: stop-cassandra
stop-cassandra:
	@docker stop cqlx-cassandra >/dev/null 2>&1 || true

.PHONY: get-deps
get-deps:
	@go mod download

.PHONY: get-tools
get-tools:
	@echo "==> Installing tools at $(GOBIN)..."
	@$(MAKE) install-golangci-lint
	@$(MAKE) install-fieldalignment

.require-golangci-lint:
ifeq ($(shell if golangci-lint --version 2>/dev/null | grep ${GOLANGCI_VERSION} 1>/dev/null 2>&1; then echo "ok"; else echo "need-install"; fi), need-install)
	$(MAKE) install-golangci-lint
endif

install-golangci-lint:
	@echo "==> Installing golangci-lint ${GOLANGCI_VERSION} at $(GOBIN)..."
	$(call dl_tgz,golangci-lint,$(GOLANGCI_DOWNLOAD_URL))

.require-fieldalignment:
ifeq ($(shell if command -v fieldalignment >/dev/null 2>&1; then echo "ok"; else echo "need-install"; fi), need-install)
	$(MAKE) install-fieldalignment
endif

install-fieldalignment:
	@echo "==> Installing fieldalignment at $(GOBIN)..."
	@go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

define dl_tgz
	@mkdir "$(GOBIN)" 2>/dev/null || true
	@echo "Downloading $(GOBIN)/$(1)";
	@curl --progress-bar -L $(2) | tar zxf - --wildcards --strip 1 -C $(GOBIN) '*/$(1)';
	@chmod +x "$(GOBIN)/$(1)";
endef
