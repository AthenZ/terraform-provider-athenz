GOPKGNAME = github.com/AthenZ/terraform-provider-athenz

export GOPATH ?= $(shell go env GOPATH)

BINARY=terraform-provider-athenz
FMT_LOG=/tmp/fmt.log
GOIMPORTS_LOG=/tmp/goimports.log

vet:
	go vet $(GOPKGNAME)/...

fmt:
	gofmt -d . >$(FMT_LOG)
	@if [ -s $(FMT_LOG) ]; then echo gofmt FAIL; cat $(FMT_LOG); false; fi

goimports:
	go install golang.org/x/tools/cmd/goimports

go_import:
	goimports -d . >$(GOIMPORTS_LOG)
	@if [ -s $(GOIMPORTS_LOG) ]; then echo goimports FAIL; cat $(GOIMPORTS_LOG); false; fi

build_mac:
	GOOS=darwin go install -v $(GOPKGNAME)/...

build_linux:
	GOOS=linux go install -v $(GOPKGNAME)/...

unit: vet fmt
	TF_ACC=false ; go test -v $(GOPKGNAME)/...

test: unit

build: test build_linux build_mac

build_no_test: build_linux build_mac
