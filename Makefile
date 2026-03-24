SHELL := /bin/bash

VERSION	= 0.6.0

TERRAFORM	  = terraform
TERRAFORM_VERSION = "~> 1.5"

GO	 = go
MODULE	 = $(shell env GO111MODULE=on $(GO) list -m)
PKGS	 = $(shell env GO111MODULE=on $(GO) list ./... | grep -v /vendor/)
TESTPKGS = $(shell env GO111MODULE=on $(GO) list -f \
		'{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' \
		$(PKGS))
TESTPKGS_MSSQL = $(filter %/mssql,$(TESTPKGS))
TESTPKGS_OTHER = $(filter-out %/mssql,$(TESTPKGS))

ifeq ($(OS),Windows_NT)
	OPERATING_SYSTEM=Windows
	ifeq ($(PROCESSOR_ARCHITEW6432),AMD64)
		OS_ARCH=windows_amd64
	else
		ifeq ($(PROCESSOR_ARCHITECTURE),AMD64)
			OS_ARCH=windows_amd64
		endif
		ifeq ($(PROCESSOR_ARCHITECTURE),x86)
			OS_ARCH=windows_386
		endif
	endif
else
	UNAME_S := $(shell uname -s)
	ifeq ($(UNAME_S),Linux)
		OPERATING_SYSTEM=Linux
		_OS=linux
	endif
	ifeq ($(UNAME_S),Darwin)
		OPERATING_SYSTEM=MacOS
		_OS=darwin
	endif
	UNAME_P := $(shell uname -p)
	ifeq ($(UNAME_P),x86_64)
		OS_ARCH=$(_OS)_amd64
	endif
	ifneq ($(filter %86,$(UNAME_P)),)
		OS_ARCH=$(_OS)_386
	endif
	ifneq ($(filter arm%,$(UNAME_P)),)
		OS_ARCH=$(_OS)_arm
	endif
endif

INSTALL_PATH=~/.terraform.d/plugins/$(shell basename $(shell dirname $(MODULE)))/$(shell basename $(MODULE) | cut -d'-' -f3)/${VERSION}/${OS_ARCH}

default: install

build:
	CGO_ENABLED=0 $(GO) build -o $(shell basename $(MODULE))

release:
	# Runs goreleaser locally (testrun)
	goreleaser release --rm-dist --skip-sign --skip-publish

install: build
	mkdir -p $(INSTALL_PATH)
	mv $(shell basename $(MODULE)) $(INSTALL_PATH)/

test:
	echo $(TESTPKGS) | xargs -t -n4 $(GO) test $(TESTARGS) -timeout=30s -parallel=4

testacc: testacc-mssql testacc-other

testacc-mssql:
	if [ -f .local.env ]; then source .local.env; fi && TF_ACC=1 TERRAFORM_VERSION=$(TERRAFORM_VERSION) $(GO) test $(TESTPKGS_MSSQL) -v $(TESTARGS) -timeout 120m

testacc-other:
	if [ -f .local.env ]; then source .local.env; fi && TF_ACC=1 TERRAFORM_VERSION=$(TERRAFORM_VERSION) $(GO) test $(TESTPKGS_OTHER) -v $(TESTARGS) -timeout 120m

testacc-local: testacc-local-mssql testacc-local-other

testacc-local-mssql:
	if [ -f .local.env ]; then source .local.env; fi && TF_ACC_LOCAL=1 TERRAFORM_VERSION=$(TERRAFORM_VERSION) $(GO) test $(TESTPKGS_MSSQL) -v $(TESTARGS) -timeout 120m

testacc-local-other:
	if [ -f .local.env ]; then source .local.env; fi && TF_ACC_LOCAL=1 TERRAFORM_VERSION=$(TERRAFORM_VERSION) $(GO) test $(TESTPKGS_OTHER) -v $(TESTARGS) -timeout 120m

docker-start:
	cd test-fixtures/local && export TERRAFORM_VERSION=$(TERRAFORM_VERSION) && ${TERRAFORM} init && ${TERRAFORM} apply -auto-approve -var="operating_system=${OPERATING_SYSTEM}"

docker-stop:
	cd test-fixtures/local && TERRAFORM_VERSION=$(TERRAFORM_VERSION) ${TERRAFORM} destroy -auto-approve -var="operating_system=${OPERATING_SYSTEM}"

azure-create:
	cd test-fixtures/all && export TERRAFORM_VERSION=$(TERRAFORM_VERSION) && ${TERRAFORM} init && ${TERRAFORM} apply -auto-approve

azure-destroy:
	cd test-fixtures/all && TERRAFORM_VERSION=$(TERRAFORM_VERSION) ${TERRAFORM} destroy -auto-approve
