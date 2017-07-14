export GOPATH := $(GOPATH):$(PWD):$(PWD)/vendor
VERSION = $(shell cat ./VERSION)

.PHONY: all checksetup glide_rename glide_install glide_reinstall glide_update check build test ci run show_version deploy update-traffic

all: check

checksetup:
	go get golang.org/x/tools/cmd/goimports

glide_rename:
	cd vendor && mv src/ vendor/

glide_install:
	cd vendor && glide install
	cd vendor && mv vendor/ src/

glide_reinstall: glide_rename glide_install

glide_update:
	cd vendor && rm -r -f src/
	cd vendor && glide update
	cd vendor && mv vendor/ src/


check:
	go fmt src/admin/*.go
	go fmt src/api/*.go
	go fmt src/gae_support/*.go
	go fmt src/models/*.go
	go fmt src/test_utils/*.go
	go fmt app/concurrent-batch-agent/*.go

	go vet src/admin/*.go
	go vet src/api/*.go
	go vet src/gae_support/*.go
	go vet src/models/*.go
	go vet src/test_utils/*.go
	go vet app/concurrent-batch-agent/*.go

	goimports -l src/admin/*.go
	goimports -l src/api/*.go
	goimports -l src/gae_support/*.go
	goimports -l src/models/*.go
	goimports -l src/test_utils/*.go
	goimports -l app/concurrent-batch-agent/*.go

	git diff --exit-code

build:
	goapp build	./src/models \
							./src/api \
							./src/admin

test:
	goapp test	./src/models \
							./src/api \
							./src/admin

ci: check test

run:
	dev_appserver.py ./app/concurrent-batch-agent/app.yaml

show_version:
	@echo ${VERSION}

deploy: build
	appcfg.py -A $${PROJECT} -V ${VERSION} update ./app/concurrent-batch-agent

update-traffic:
	gcloud --project ${PROJECT} app services set-traffic concurrent-batch-agent --splits=${VERSION}=1 -q
