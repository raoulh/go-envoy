GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=envoy
VERSION?=0.0.0

GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

.PHONY: all test build install

all: help

## Help:
help: ## Show this help.
        @echo ''
        @echo 'Usage:'
        @echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
        @echo ''
        @echo 'Targets:'
        @awk 'BEGIN {FS = ":.*?## "} { \
                if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
                else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
                }' $(MAKEFILE_LIST)

## Build:
build: build-server build-cli ## Build the project and put the output binary in out

build-server: ## Build only the envoy server
        @mkdir -p out
		@cd cmd/web
        $(GOCMD) build -v -o ../../out/envoy_web .

build-server: ## Build only the CLI tool
        @mkdir -p out
		@cd cmd/envoy
        $(GOCMD) build -v -o ../../out/envoy .

build-www: ## Build the www files
		@mkdir -p out
		@cp -R web out

clean: ## Remove build related file
        rm -fr ./bin
        rm -fr ./out
        rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

## Install:
install: install-bins install-data ## Install the project

install-bins:
		install -d $(DESTDIR)$(PREFIX)/bin/
		install -m 755 out/envoy $(DESTDIR)$(PREFIX)/bin/
		install -m 755 out/envoy_web $(DESTDIR)$(PREFIX)/bin/

install-data:
		install -d $(DESTDIR)$(PREFIX)/share/envoy
		install -d $(DESTDIR)$(PREFIX)/share/envoy/css
		install -d $(DESTDIR)$(PREFIX)/share/envoy/js
		install -d $(DESTDIR)$(PREFIX)/share/envoy/templates
		install -m 644 out/web/css/simple.css $(DESTDIR)$(PREFIX)/share/envoy/css
		install -m 644 out/web/js/index.js $(DESTDIR)$(PREFIX)/share/envoy/js
		install -m 644 out/web/templates/index.html $(DESTDIR)$(PREFIX)/share/envoy/templates

## Test:
test: ## Run the tests of the project
        $(GOTEST) -v -race ./... $(OUTPUT_OPTIONS)

coverage: ## Run the tests of the project and export the coverage
        $(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
        $(GOCMD) tool cover -func profile.cov