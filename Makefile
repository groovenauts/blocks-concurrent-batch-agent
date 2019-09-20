REPO:=github.com/groovenauts/blocks-concurrent-batch-server

DEPLOY_ENV ?= staging

SERVICE_NAME=concurrent-batch

VERSION ?= $(shell cat ./VERSION)

BASE_PACKAGE_PATH=$(REPO)
APP_PACKAGE_PATH=$(REPO)/app/concurrent-batch-agent
TEST_PACKAGES=$(SERVERBASE_PACKAGE_PATH)/... $(BASE_PACKAGE_PATH)/ $(BASE_PACKAGE_PATH)/scenario_tests/

.PHONY: dep_ensure
dep_ensure:
	dep ensure

.PHONY: dep_update
dep_update:
	dep ensure -update

vendor:
	dep ensure -vendor-only

.PHONY: build
build: vendor
	mkdir -p tmp/ && \
	go build -o tmp/build $(APP_PACKAGE_PATH)

.PHONY: test
test: vendor
	go test $(BASE_PACKAGE_PATH)/src/...
