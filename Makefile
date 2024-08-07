GOPKGNAME = github.com/AthenZ/terraform-provider-athenz

export GOPATH ?= $(shell /usr/local/go/bin/go env GOPATH)

BINARY=terraform-provider-athenz
FMT_LOG=/tmp/fmt.log
GOIMPORTS_LOG=/tmp/goimports.log

# define go build params for local terrafrom build
GOOS    := $(shell uname -s | tr '[A-Z]' '[a-z]' )
GOARCH  := $(shell uname -m | sed s/x86_64/amd64/ )
OS_ARCH := $(GOOS)_$(GOARCH)
VERSION := 9.9.9

ifndef SYS_TEST_CA_CERT
	SYS_TEST_CA_CERT=$(shell pwd)/docker/sample/CAs/athenz_ca.pem 
endif
ifndef SYS_TEST_CERT
	SYS_TEST_CERT=$(shell pwd)/docker/sample/domain-admin/domain_admin_cert.pem
endif
ifndef SYS_TEST_KEY
	SYS_TEST_KEY=$(shell pwd)/docker/sample/domain-admin/domain_admin_key.pem
endif

vet:
	/usr/local/go/bin/go vet $(GOPKGNAME)/...

fmt:
	/usr/local/go/bin/gofmt -d . >$(FMT_LOG)
	@if [ -s $(FMT_LOG) ]; then echo gofmt FAIL; cat $(FMT_LOG); false; fi

goimports:
	/usr/local/go/bin/go install golang.org/x/tools/cmd/goimports

go_import:
	goimports -d . >$(GOIMPORTS_LOG)
	@if [ -s $(GOIMPORTS_LOG) ]; then echo goimports FAIL; cat $(GOIMPORTS_LOG); false; fi

build_mac:
	GOOS=darwin go install -v $(GOPKGNAME)/...

build_linux:
	GOOS=linux go install -v $(GOPKGNAME)/...

install_local:
	@echo VERSION: $(VERSION) OS_ARCH: $(OS_ARCH) GOOS: $(GOOS) GOARCH: $(GOARCH)
	/usr/local/go/bin/go build -o ${BINARY}
	mkdir -p ~/.terraform.d/plugins/yahoo/provider/athenz/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/yahoo/provider/athenz/${VERSION}/${OS_ARCH}

unit: vet fmt
	export TF_ACC=false ; /usr/local/go/bin/go test -v $(GOPKGNAME)/...

acc_test: vet fmt
	@echo acc_test: cacert: $(SYS_TEST_CA_CERT)
	@echo acc_test: cert: $(SYS_TEST_CERT)
	@echo acc_test: key: $(SYS_TEST_KEY)
	export MEMBER_1=terraform-provider.athenz_provider_foo MEMBER_2=user.github-7654321 ADMIN_USER=user.github-7654321 SHORT_ID=github-7654321 TOP_LEVEL_DOMAIN=terraformTest DELEGATED_DOMAIN=terraform-provider-delegate DOMAIN=terraform-provider PARENT_DOMAIN=terraform-provider SUB_DOMAIN=Test DOMAIN=terraform-provider export TF_ACC=true export ATHENZ_CA_CERT=$(SYS_TEST_CA_CERT) export ATHENZ_ZMS_URL=https://localhost:4443/zms/v1 export ATHENZ_CERT=$(SYS_TEST_CERT) export ATHENZ_KEY=$(SYS_TEST_KEY) ; /usr/local/go/bin/go test -v $(GOPKGNAME)/...

test: unit

build: test build_linux build_mac

build_no_test: build_linux build_mac
