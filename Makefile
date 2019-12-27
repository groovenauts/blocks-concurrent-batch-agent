REPO:=github.com/groovenauts/blocks-concurrent-batch-agent

SERVICE_NAME=concurrent-batch-agent

VERSION ?= $(shell cat ./VERSION)

GOPATH=$(shell go env GOPATH)
BASE_PACKAGE_PATH=$(REPO)
APP_PATH=app/concurrent-batch-agent
APP_PACKAGE_PATH=$(REPO)/$(APP_PATH)
TEST_PACKAGES=$(SERVERBASE_PACKAGE_PATH)/... $(BASE_PACKAGE_PATH)/ $(BASE_PACKAGE_PATH)/scenario_tests/

APP_YAML_PATH=$(APP_PATH)/app.yaml

tmp:
	mkdir -p tmp

.PHONY: build
build: tmp
	go build -o tmp/build $(APP_PACKAGE_PATH)

.PHONY: test
test:
	go test $(BASE_PACKAGE_PATH)/src/...

.PHONY: GOPATH
GOPATH:
	@go env GOPATH

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: ci
ci:	test

$(APP_YAML_PATH):
	erb -T - $(APP_YAML_PATH).erb > $(APP_YAML_PATH)

.PHONY: deploy
deploy: build $(APP_YAML_PATH)
	gcloud --project=$(PROJECT) app deploy $(APP_YAML_PATH) --version=${VERSION} --no-promote --quiet

.PHONY: update-traffic
update-traffic:
	gcloud --project=$(PROJECT) app services set-traffic $(SERVICE_NAME) --splits=${VERSION}=1 -q
