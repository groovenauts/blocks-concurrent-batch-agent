# See https://github.com/vincentbernat/hellogopher/blob/master/Makefile
GITHUB_ORG  = groovenauts
GITHUB_REPO = blocks-concurrent-batch-agent
PACKAGE  = concurrent-batch-agent
DATE    ?= $(shell date +%FT%T%z)
VERSION ?= $(shell cat ./VERSION)
GOPATH   = $(CURDIR)/.gopath~
BIN      = $(GOPATH)/bin
GOPATH_SRC=$(GOPATH)/src
BASE     = $(GOPATH)/src/$(PACKAGE)
PKGS     = $(or $(PKG),$(shell cd $(BASE) && env GOPATH=$(GOPATH) $(GO) list ./... | grep -v "^$(PACKAGE)/vendor/"))
TESTPKGS = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
PKGDIR   = $(CURDIR)/pkg
SRC_DIRS = $(subst /,, $(subst src/,,$(sort $(dir $(wildcard src/*/)))))
MAIN_PACKAGES=models api admin

.PHONY: envs
envs:
	@echo "SRC_DIRS: $(SRC_DIRS)"

export GOPATH

GO      = go
GOAPP   = goapp
GODOC   = godoc
GOFMT   = gofmt
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m▶\033[0m")

.PHONY: all
all: build

$(GOPATH_SRC): ; $(info $(M) setting GOPATH…)
	@mkdir -p $@
	for sd in $(SRC_DIRS); do \
		ln -sf $(CURDIR)/src/$$sd $@/$$sd ;\
	done

$(BASE): $(GOPATH_SRC)
	@mkdir -p $(dir $@)
	@ln -sf $(CURDIR) $@

# Tools

$(BIN):
	@mkdir -p $@
$(BIN)/%: $(BIN) | $(BASE) ; $(info $(M) building $(REPOSITORY)…)
	$Q tmp=$$(mktemp -d); \
		(GOPATH=$$tmp go get $(REPOSITORY) && cp $$tmp/bin/* $(BIN)/.) || ret=$$?; \
		rm -rf $$tmp ; exit $$ret

GODEP = $(BIN)/dep
$(BIN)/dep: REPOSITORY=github.com/golang/dep/cmd/dep

GOLINT = $(BIN)/golint
$(BIN)/golint: REPOSITORY=github.com/golang/lint/golint

GOCOVMERGE = $(BIN)/gocovmerge
$(BIN)/gocovmerge: REPOSITORY=github.com/wadey/gocovmerge

GOCOV = $(BIN)/gocov
$(BIN)/gocov: REPOSITORY=github.com/axw/gocov/...

GOCOVXML = $(BIN)/gocov-xml
$(BIN)/gocov-xml: REPOSITORY=github.com/AlekSi/gocov-xml

GO2XUNIT = $(BIN)/go2xunit
$(BIN)/go2xunit: REPOSITORY=github.com/tebeka/go2xunit

GOX = $(BIN)/gox
$(BIN)/gox: REPOSITORY=github.com/mitchellh/gox

GHR = $(BIN)/ghr
$(BIN)/ghr: REPOSITORY=github.com/tcnksm/ghr

## Server

.PHONY: build
build: fmt vendor | $(BASE) ; $(info $(M) building executable…) @ ## Build program binary
	$Q cd $(BASE) && $(GOAPP) build $(MAIN_PACKAGES)


.PHONY: run
run: fmt vendor | $(BASE) ; $(info $(M) Running dev server…) @ ## Running dev_appserver
	$Q dev_appserver.py $(CURDIR)/app/concurrent-batch-agent/app.yaml


# Tests

TEST_TARGETS := test-models test-api test-admin
.PHONY: $(TEST_TARGETS) test _test
$(TEST_TARGETS): NAME=$(MAKECMDGOALS:test-%=%)
$(TEST_TARGETS): ARGS=$(NAME)
$(TEST_TARGETS): _test
test: ARGS=$(MAIN_PACKAGES)
test:	_test
_test: fmt vendor | $(BASE) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests
		$Q cd $(BASE) && $(GOAPP) test $(ARGS)

test-xml: fmt vendor | $(BASE) $(GO2XUNIT) ; $(info $(M) running $(NAME:%=% )tests…) @ ## Run tests with xUnit output
	$Q cd $(BASE) && 2>&1 $(GO) test -timeout 20s -v $(TESTPKGS) | tee test/tests.output
	$(GO2XUNIT) -fail -input test/tests.output -output test/tests.xml

COVERAGE_MODE = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML = $(COVERAGE_DIR)/index.html
.PHONY: test-coverage test-coverage-tools
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML)
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt vendor test-coverage-tools | $(BASE) ; $(info $(M) running coverage tests…) @ ## Run coverage tests
	$Q mkdir -p $(COVERAGE_DIR)/coverage
	$Q cd $(BASE) && for pkg in $(MAIN_PACKAGES); do \
		$(GOAPP) test \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	 done
	$Q $(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	$Q $(GO) tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	$Q $(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)

.PHONY: lint
lint: vendor | $(BASE) $(GOLINT) ; $(info $(M) running golint…) @ ## Run golint
	$Q cd $(BASE) && ret=0 && for pkg in $(PKGS); do \
		test -z "$$($(GOLINT) $$pkg | tee /dev/stderr)" || ret=1 ; \
	 done ; exit $$ret

.PHONY: fmt
fmt: ; $(info $(M) running gofmt…) @ ## Run gofmt on all source files
	@ret=0 && for d in $$($(GO) list -f '{{.Dir}}' ./... | grep -v /vendor/); do \
		$(GOFMT) -l -w $$d/*.go || ret=$$? ; \
	 done ; exit $$ret

# Dependency management
.PHONY: dep_ensure
dep_ensure: ; $(info $(M) retrieving dependencies…)
	@cd $(BASE) && $(GODEP) ensure

vendor: Gopkg.toml Gopkg.lock | $(BASE) $(GODEP) dep_ensure
	@rm -rf $(CURDIR)/vendor/src
	@cd $(BASE) && $(GODEP) ensure
	$Q for d in $(subst /,, $(subst vendor/,, $(sort $(dir $(wildcard vendor/*/))))); do \
	  cd $(GOPATH_SRC) && ln -sf $(CURDIR)/vendor/$$d $$d ;\
	done
	@ln -nsf . vendor/src
	@touch $@
.PHONY: vendor-init
vendor-init: $(BASE)
	cd $(BASE) && dep init
.PHONY: vendor-update
vendor-update: vendor | $(BASE) $(GODEP)
ifeq "$(origin PKG)" "command line"
	$(info $(M) updating $(PKG) dependency…)
	$Q cd $(BASE) && $(GODEP) ensure -update $(PKG)
else
	$(info $(M) updating all dependencies…)
	$Q cd $(BASE) && $(GODEP) ensure -update
endif
	@ln -nsf . vendor/src
	@touch vendor

# Misc

.PHONY: clean
clean: ; $(info $(M) cleaning…)	@ ## Cleanup everything
	@rm -rf $(GOPATH)
	@rm -rf bin
	@rm -rf test/tests.* test/coverage.*

.PHONY: help
help:
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: git_guard git_tag git_push_tag tag
git_guard:
	$Q git diff --exit-code
git_tag:
	git tag v${VERSION}
git_push_tag:
	git push origin v${VERSION}
tag: git_tag git_push_tag


.PHONY: ci
ci:	fmt git_guard test

.PHONY: deploy
deploy: build
	appcfg.py -A $${PROJECT} -V ${VERSION} update ./app/concurrent-batch-agent

.PHONY: update-traffic
update-traffic:
	gcloud --project ${PROJECT} app services set-traffic concurrent-batch-agent --splits=${VERSION}=1 -q
