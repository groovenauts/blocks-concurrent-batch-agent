UNFORMATTED=$(shell gofmt -l *.go)

all: check

checksetup:
	go get golang.org/x/tools/cmd/goimports

check: checkfmt
	go vet *.go
	goimports -l *.go

checkfmt:
ifneq ($(UNFORMATTED),)
	@echo $(UNFORMATTED)
	exit 1
else
	@echo "gofmt -l *.go OK"
endif

test:
	goapp test github.com/groovenauts/blocks-concurrent-batch-agent/models
	goapp test
