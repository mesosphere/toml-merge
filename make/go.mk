# Copyright 2022 D2iQ, Inc. All rights reserved.
# SPDX-License-Identifier: Apache-2.0

# The GOPRIVATE environment variable controls which modules the go command considers
# to be private (not available publicly) and should therefore not use the proxy or checksum database
export GOPRIVATE ?=

ALL_GO_SUBMODULES := $(shell PATH='$(PATH)'; find -mindepth 2 -maxdepth 2 -name go.mod -printf '%P\n' | sort)
GO_SUBMODULES_NO_TOOLS := $(filter-out $(addsuffix /go.mod,tools),$(ALL_GO_SUBMODULES))

ifndef GOOS
export GOOS := $(OS)
endif
ifndef GOARCH
export GOARCH := $(shell go env GOARCH)
endif

define go_test
	gotestsum \
		--jsonfile test.json \
		--junitfile junit-report.xml \
		--junitfile-testsuite-name=relative \
		--junitfile-testcase-classname=short \
		-- \
		-covermode=atomic \
		-coverprofile=coverage.out \
		-race \
		-short \
		-v \
		$(if $(GOTEST_RUN),-run "$(GOTEST_RUN)") \
		./... && \
	go tool cover \
		-html=coverage.out \
		-o coverage.html
endef

.PHONY: test
test: ## Runs go tests for all modules in repository
ifneq ($(wildcard $(REPO_ROOT)/go.mod),)
test: test.root
endif
ifneq ($(words $(GO_SUBMODULES_NO_TOOLS)),0)
test: $(addprefix test.,$(GO_SUBMODULES_NO_TOOLS:/go.mod=))
endif

.PHONY: test.%
test.%: ## Runs go tests for a specific module
test.%: install-tool.go.gotestsum; $(info $(M) running tests$(if $(GOTEST_RUN), matching "$(GOTEST_RUN)") for $* module)
	$(if $(filter-out root,$*),cd $* && )$(call go_test)

.PHONY: integration-test
integration-test: ## Runs integration tests for all modules in repository
integration-test:
	$(MAKE) GOTEST_RUN=Integration test

.PHONY: integration-test.%
integration-test.%: ## Runs integration tests for a specific module
integration-test.%:
	$(MAKE) GOTEST_RUN=Integration test.$*

.PHONY: bench
bench: ## Runs go benchmarks for all modules in repository
ifneq ($(wildcard $(REPO_ROOT)/go.mod),)
bench: bench.root
endif
ifneq ($(words $(GO_SUBMODULES_NO_TOOLS)),0)
bench: $(addprefix bench.,$(GO_SUBMODULES_NO_TOOLS:/go.mod=))
endif

.PHONY: bench.%
bench.%: ## Runs go benchmarks for a specific module
bench.%: ; $(info $(M) running benchmarks$(if $(GOTEST_RUN), matching "$(GOTEST_RUN)") for $* module)
	$(if $(filter-out root,$*),cd $* && )go test $(if $(GOTEST_RUN),-run "$(GOTEST_RUN)") -race -cover -v ./...

E2E_PARALLEL_NODES ?= $(shell nproc --ignore=1)
E2E_FLAKE_ATTEMPTS ?= 1

.PHONY: e2e-test
e2e-test: ## Runs e2e tests
e2e-test: install-tool.golang install-tool.ginkgo build-snapshot
	$(info $(M) running e2e tests$(if $(E2E_LABEL), labelled "$(E2E_LABEL)")$(if $(E2E_FOCUS), matching "$(E2E_FOCUS)"))
	ginkgo run \
		--r \
		--race \
		--show-node-events \
		--trace \
		--randomize-all \
		--randomize-suites \
		--fail-on-pending \
		--keep-going \
		$(if $(filter $(CI),true),--always-emit-ginkgo-writer) \
		--covermode=atomic \
		--coverprofile coverage-e2e.out \
		--procs=$(E2E_PARALLEL_NODES) \
		--compilers=$(E2E_PARALLEL_NODES) \
		--flake-attempts=$(E2E_FLAKE_ATTEMPTS) \
		$(if $(E2E_FOCUS),--focus="$(E2E_FOCUS)") \
		$(if $(E2E_SKIP),--skip="$(E2E_SKIP)") \
		$(if $(E2E_LABEL),--label-filter="$(E2E_LABEL)") \
		$(E2E_GINKGO_FLAGS) \
		--junit-report=junit-e2e.xml \
		--json-report=report-e2e.json \
		--tags e2e \
		test/e2e/... && \
	go tool cover \
		-html=coverage-e2e.out \
		-o coverage-e2e.html

GOLANGCI_CONFIG_FILE ?= $(wildcard $(REPO_ROOT)/.golangci.y*ml)

.PHONY: lint
lint: ## Runs golangci-lint for all modules in repository
ifneq ($(wildcard $(REPO_ROOT)/go.mod),)
lint: lint.root
endif
ifneq ($(words $(GO_SUBMODULES_NO_TOOLS)),0)
lint: $(addprefix lint.,$(GO_SUBMODULES_NO_TOOLS:/go.mod=))
endif

.PHONY: lint.%
lint.%: ## Runs golangci-lint for a specific module
lint.%: install-tool.golangci-lint install-tool.go.golines; $(info $(M) linting $* module)
	$(if $(filter-out root,$*),cd $* && )golines -w .
	$(if $(filter-out root,$*),cd $* && )golangci-lint run --fix --config=$(GOLANGCI_CONFIG_FILE)
	$(if $(filter-out root,$*),cd $* && )go fix ./...

.PHONY: mod-tidy
mod-tidy: ## Run go mod tidy for all modules
ifneq ($(wildcard $(REPO_ROOT)/go.mod),)
mod-tidy: mod-tidy.root
endif
ifneq ($(words $(ALL_GO_SUBMODULES)),0)
mod-tidy: $(addprefix mod-tidy.,$(ALL_GO_SUBMODULES:/go.mod=))
endif

.PHONY: mod-tidy.%
mod-tidy.%: ## Runs go mod tidy for a specific module
mod-tidy.%: install-tool.golang; $(info $(M) running go mod tidy for $* module)
	$(if $(filter-out root,$*),cd $* && )go mod tidy -v -compat=1.17
	$(if $(filter-out root,$*),cd $* && )go mod verify

.PHONY: go-clean
go-clean: ## Cleans go build, test and modules caches for all modules
ifneq ($(wildcard $(REPO_ROOT)/go.mod),)
go-clean: go-clean.root
endif
ifneq ($(words $(ALL_GO_SUBMODULES)),0)
go-clean: $(addprefix go-clean.,$(ALL_GO_SUBMODULES:/go.mod=))
endif

.PHONY: go-clean.%
go-clean.%: ## Cleans go build, test and modules caches for a specific module
go-clean.%: install-tool.golang; $(info $(M) running go clean for $* module)
	$(if $(filter-out root,$*),cd $* && )go clean -r -i -cache -testcache -modcache

.PHONY: go-generate
go-generate: ## Runs go generate
go-generate: install-tool.golang ; $(info $(M) running go generate)
	go generate -x ./...

.PHONY: go-mod-upgrade
go-mod-upgrade: ## Interactive check for direct module dependency upgrades
go-mod-upgrade: install-tool.go.go-mod-upgrade; $(info $(M) checking for direct module dependency upgrades)
	go-mod-upgrade
