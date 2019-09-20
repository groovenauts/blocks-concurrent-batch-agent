REPO:=github.com/groovenauts/blocks-concurrent-batch-server

SERVICE_NAME=concurrent-batch-agent

VERSION ?= $(shell cat ./VERSION)

BASE_PACKAGE_PATH=$(REPO)
APP_PATH=app/concurrent-batch-agent
APP_PACKAGE_PATH=$(REPO)/$(APP_PATH)
TEST_PACKAGES=$(SERVERBASE_PACKAGE_PATH)/... $(BASE_PACKAGE_PATH)/ $(BASE_PACKAGE_PATH)/scenario_tests/

APP_YAML_PATH=$(APP_PATH)/app.yaml

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
